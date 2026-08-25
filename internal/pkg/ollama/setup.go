/**
 * Copyright 2026 AiT authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

package ollama

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// Options configures setup and EnsureLocal.
type Options struct {
	Model      string
	BaseURL    string
	ModelsDir  string // --models-dir flag override
	AssumeYes  bool
	NoInstall  bool
	NoPull     bool
	DryRun     bool
	Out        io.Writer
	Err        io.Writer
	In         io.Reader
	Runner     CommandRunner
	Confirm    ConfirmFunc
	HTTPClient *http.Client
}

// Result summarizes setup steps.
type Result struct {
	ConfigWritten   bool
	InstallSkipped  bool
	InstallAttempt  bool
	InstallOK       bool
	APIOK           bool
	ModelPulled     bool
	ManualHint      bool
	DeclinedInstall bool
}

// RunSetup applies tiered config and ensures local Ollama when possible.
func RunSetup(ctx context.Context, cfg *config.Config, opts Options) (Result, error) {
	opts = opts.withDefaults()
	var res Result

	if opts.DryRun {
		printDryRun(opts, cfg)
		return res, nil
	}

	in := opts.In
	if in == nil {
		in = os.Stdin
	}
	modelsDir, err := ChooseModelsDir(in, opts.Out, cfg, opts.AssumeYes, opts.ModelsDir)
	if err != nil {
		return res, err
	}
	logStep(opts.Out, 1, 6, fmt.Sprintf("Models directory: %s", modelsDir))

	if err := ApplyTieredConfig(cfg, opts.Model, opts.BaseURL); err != nil {
		return res, err
	}
	res.ConfigWritten = true
	logStep(opts.Out, 2, 6, "Wrote tiered + ollama to config (local-first routing enabled)")

	ensureRes, err := EnsureLocal(ctx, cfg, opts)
	res = mergeResult(res, ensureRes)
	if err != nil && !res.DeclinedInstall {
		return res, err
	}

	printNext(opts.Out)
	return res, nil
}

// EnsureLocal checks availability and optionally installs/starts/pulls Ollama.
func EnsureLocal(ctx context.Context, cfg *config.Config, opts Options) (Result, error) {
	opts = opts.withDefaults()
	var res Result

	baseURL := opts.BaseURL
	model := opts.Model
	if cfg != nil {
		if pc, ok := cfg.Provider(config.ProviderOllama); ok {
			if pc.BaseURL != "" {
				baseURL = pc.BaseURL
			}
			if pc.Model != "" {
				model = pc.Model
			}
		}
	}
	if baseURL == "" {
		baseURL = config.DefaultOllamaBase
	}
	if model == "" {
		model = config.DefaultOllamaModel
	}

	modelsDir := ModelsDirFromConfig(cfg)
	if modelsDir == "" {
		var err error
		modelsDir, err = ResolveModelsDir(cfg, opts.ModelsDir)
		if err != nil {
			return res, err
		}
	}

	avail := CheckAvailable(ctx, opts.HTTPClient, baseURL)
	if avail.BinaryOK && avail.APIOK {
		res.APIOK = true
		logStep(opts.Out, 3, 6, "Ollama binary: ok")
		logStep(opts.Out, 4, 6, logStepAPIReachable)
		if err := EnsureModel(ctx, opts.HTTPClient, baseURL, model, opts.NoPull, modelsDir); err != nil {
			return res, err
		}
		if !opts.NoPull {
			res.ModelPulled = true
			logStep(opts.Out, 5, 6, fmt.Sprintf("Model %s ready", model))
		}
		return res, nil
	}

	binaryMissing := !avail.BinaryOK
	allowed, err := ResolveInstallAllowed(binaryMissing, opts, opts.Confirm)
	if err != nil {
		return res, err
	}
	if binaryMissing && !allowed {
		res.InstallSkipped = true
		res.DeclinedInstall = true
		res.ManualHint = true
		logStep(opts.Out, 3, 6, "Ollama binary: install skipped (user declined)")
		ManualSetupInstructions(opts.Out, cfg)
		tryProbeWithoutInstall(ctx, opts, baseURL, model, modelsDir, &res)
		return res, nil
	}

	if allowed {
		res.InstallAttempt = true
		logStep(opts.Out, 3, 6, "Installing Ollama...")
		if err := EnsureBinary(ctx, opts.Runner, true); err != nil {
			ManualSetupInstructions(opts.Out, cfg)
			return res, err
		}
		res.InstallOK = true
		logStep(opts.Out, 3, 6, "Ollama binary: installed")
	}

	if err := WaitReady(ctx, opts.HTTPClient, baseURL, modelsDir); err != nil {
		ManualSetupInstructions(opts.Out, cfg)
		logStep(opts.Out, 4, 6, "API: not reachable — "+err.Error())
		return res, err
	}
	res.APIOK = true
	logStep(opts.Out, 4, 6, logStepAPIReachable)

	if err := EnsureModel(ctx, opts.HTTPClient, baseURL, model, opts.NoPull, modelsDir); err != nil {
		return res, err
	}
	if !opts.NoPull {
		res.ModelPulled = true
		logStep(opts.Out, 5, 6, fmt.Sprintf("Model %s ready", model))
	} else {
		logStep(opts.Out, 5, 6, fmt.Sprintf("Model pull skipped (run: ollama pull %s)", model))
	}
	return res, nil
}

func tryProbeWithoutInstall(ctx context.Context, opts Options, baseURL, model, modelsDir string, res *Result) {
	if _, err := platform.LookPath(binaryName); err != nil {
		logStep(opts.Out, 4, 6, "API: not reachable — install/start Ollama manually")
		logStep(opts.Out, 5, 6, "Model pull: skipped")
		return
	}
	if err := WaitReady(ctx, opts.HTTPClient, baseURL, modelsDir); err != nil {
		logStep(opts.Out, 4, 6, "API: not reachable — start Ollama (ollama serve)")
		return
	}
	res.APIOK = true
	logStep(opts.Out, 4, 6, logStepAPIReachable)
	_ = EnsureModel(ctx, opts.HTTPClient, baseURL, model, opts.NoPull, modelsDir)
}

func (o Options) withDefaults() Options {
	if o.Out == nil {
		o.Out = os.Stdout
	}
	if o.Err == nil {
		o.Err = os.Stderr
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: httpClientTimeout}
	}
	return o
}

func mergeResult(a, b Result) Result {
	a.InstallSkipped = a.InstallSkipped || b.InstallSkipped
	a.InstallAttempt = a.InstallAttempt || b.InstallAttempt
	a.InstallOK = a.InstallOK || b.InstallOK
	a.APIOK = a.APIOK || b.APIOK
	a.ModelPulled = a.ModelPulled || b.ModelPulled
	a.ManualHint = a.ManualHint || b.ManualHint
	a.DeclinedInstall = a.DeclinedInstall || b.DeclinedInstall
	return a
}

func logStep(w io.Writer, n, total int, msg string) {
	if w != nil {
		_, _ = fmt.Fprintf(w, "[%d/%d] %s\n", n, total, msg)
	}
}

func printNext(w io.Writer) {
	if w == nil {
		return
	}
	_, _ = fmt.Fprintln(w, "Next: "+toolname.AICLI("")+"   (simple questions prefer local Ollama when reachable)")
	_, _ = fmt.Fprintln(w, "      "+toolname.AICLI("model list"))
}

func printDryRun(opts Options, cfg *config.Config) {
	dir, _ := ResolveModelsDir(cfg, opts.ModelsDir)
	logStep(opts.Out, 0, 0, "dry-run: models_dir="+dir)
	logStep(opts.Out, 0, 0, "dry-run: would enable ai.tiered and providers.ollama")
	if _, err := platform.LookPath(binaryName); err != nil {
		logStep(opts.Out, 0, 0, "dry-run: "+installPromptSummary()+" (unless --no-install)")
	}
}
