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

package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/os-artificer/AiT/internal/pkg/ai/provider"
	"github.com/os-artificer/AiT/internal/pkg/ai/router"
	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/creds"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/platform"

	openai "github.com/sashabaranov/go-openai"
)

// PromptMeta carries factual provider routing fields injected into the system prompt.
type PromptMeta struct {
	Provider ProviderID
	Model    string
	Mode     string
}

// Client routes chat requests across configured AI providers with failover.
type Client struct {
	cfg             *config.Config
	endpoints       map[ProviderID]*endpoint
	sessionProvider ProviderID
	initOverride    ProviderID
	skipTiered      bool
	httpClient      *http.Client
}

// ClientOption configures NewClient.
type ClientOption func(*clientOpts)

type clientOpts struct {
	providerOverride ProviderID
	httpClient       *http.Client
}

// WithProviderOverride sets the primary cloud provider for this client instance (ait ai --model).
func WithProviderOverride(id ProviderID) ClientOption {
	return func(o *clientOpts) {
		o.providerOverride = id
	}
}

// WithHTTPClient sets the HTTP client used for provider requests (tests).
func WithHTTPClient(c *http.Client) ClientOption {
	return func(o *clientOpts) {
		o.httpClient = c
	}
}

// NewClient builds a multi-provider AI client from cfg and keyring API keys.
func NewClient(cfg *config.Config, opts ...ClientOption) (*Client, error) {
	var o clientOpts
	for _, fn := range opts {
		fn(&o)
	}

	eps := make(map[ProviderID]*endpoint)

	for _, id := range config.CloudProviderIDs() {
		if !cfg.ProviderEnabled(id) {
			continue
		}
		key, err := creds.GetProviderAPIKey(id)
		if err != nil {
			continue
		}
		pc, ok := cfg.Provider(id)
		if !ok {
			pc = config.ProviderConfig{}
		}
		eps[id] = newEndpoint(id, pc, key, o.httpClient)
	}

	if cfg.TieredEnabled() {
		localID := cfg.LocalProviderID()
		if cfg.ProviderEnabled(localID) {
			if ep := initTieredEndpoint(cfg, localID, o.httpClient); ep != nil {
				eps[localID] = ep
			}
		}
	}

	if len(eps) == 0 {
		return nil, errhint.WithTryMsg("no enabled AI providers configured",
			hintNoProvidersConfigured,
			provider.ConfigSetAPIKeyCmd(ProviderDeepseek),
			provider.ProviderEnableCmd(ProviderOllama),
			provider.HintAIHelp,
		)
	}

	if !hasCloudEndpoint(eps) && !cfg.TieredEnabled() {
		return nil, errhint.WithTryMsg("no cloud AI API keys configured",
			provider.HintSetCloudAPIKey,
			provider.ConfigSetAPIKeyCmd(cfg.DefaultProviderID()),
			provider.HintOllamaSetup,
			provider.HintAIHelp,
		)
	}

	if o.providerOverride != "" {
		if !config.IsCloudProvider(o.providerOverride) {
			return nil, fmt.Errorf(errProviderOverrideCloudFmt, ProviderDeepseek, ProviderMinimax)
		}
		if !cfg.ProviderEnabled(o.providerOverride) {
			return nil, ProviderDisabledErr(string(o.providerOverride))
		}
		if _, ok := eps[o.providerOverride]; !ok {
			return nil, errhint.WithTryMsg(fmt.Sprintf(errAPIKeyNotSetFmt, o.providerOverride),
				provider.HintStoreAPIKeyInKeyring,
				provider.ConfigSetAPIKeyCmd(string(o.providerOverride)),
				provider.ConfigGetAPIKeyCmd(string(o.providerOverride)),
			)
		}
	} else if cloud := cfg.DefaultProviderID(); hasCloudEndpoint(eps) {
		if _, ok := eps[cloud]; !ok {
			return nil, fmt.Errorf(errAPIKeyNotSetWithHintFmt, cloud, router.APIKeyHint(cloud))
		}
	}

	return &Client{
		cfg:             cfg,
		endpoints:       eps,
		initOverride:    o.providerOverride,
		sessionProvider: o.providerOverride,
		skipTiered:      o.providerOverride != "",
		httpClient:      o.httpClient,
	}, nil
}

// SetSessionProvider overrides the primary cloud provider for subsequent Chat/Assist calls.
func (c *Client) SetSessionProvider(id ProviderID) error {
	if id == "" {
		c.sessionProvider = c.initOverride
		c.skipTiered = c.initOverride != ""
		return nil
	}
	if id == ProviderOllama {
		return fmt.Errorf("%s", errCannotSetOllamaSession)
	}
	if !config.IsCloudProvider(id) {
		return fmt.Errorf(errUnknownCloudProviderFmt, id)
	}
	if !c.cfg.ProviderEnabled(id) {
		return fmt.Errorf(errProviderDisabledFmt, id, id)
	}
	if _, ok := c.endpoints[id]; !ok {
		return fmt.Errorf(errProviderNoAPIKeyFmt, id, router.APIKeyHint(id))
	}
	c.sessionProvider = id
	c.skipTiered = true
	return nil
}

// SessionProvider returns the active session override, or empty if using config default.
func (c *Client) SessionProvider() ProviderID {
	return c.sessionProvider
}

// DefaultProviderID returns ai.default from config.
func (c *Client) DefaultProviderID() ProviderID {
	return c.cfg.DefaultProviderID()
}

// TieredEnabled reports whether tiered local routing is on and Ollama endpoint exists.
func (c *Client) TieredEnabled() bool {
	if c == nil || !c.cfg.TieredEnabled() {
		return false
	}
	_, ok := c.endpoints[c.cfg.LocalProviderID()]
	return ok
}

func (c *Client) completion(
	ctx context.Context, user string, historyLen int,
	build func(PromptMeta) []openai.ChatCompletionMessage,
) (string, string, error) {
	useLocal := router.ClassifyPrompt(c.cfg, user, historyLen, c.sessionProvider, c.skipTiered)
	localID := c.cfg.LocalProviderID()

	if useLocal {
		if ep, ok := c.endpoints[localID]; ok {
			meta := PromptMeta{Provider: localID, Model: ep.model, Mode: aiModeLocalFirst}
			reply, thinking, err := c.complete(ctx, ep, build(meta))
			if err == nil {
				return reply, thinking, nil
			}
			if order := router.ResolveTryOrder(c.cfg, c.sessionProvider, cloudKeyedSet(c.endpoints)); len(order) > 0 {
				router.LogFailover(nil, localID, order[0], err)
			}
		}
	}

	return c.chatWithFailover(ctx, build)
}

func (c *Client) completionAgent(
	ctx context.Context, user string, historyLen int,
	build func(PromptMeta) []openai.ChatCompletionMessage,
) (string, string, error) {
	useLocal := router.ClassifyPrompt(c.cfg, user, historyLen, c.sessionProvider, c.skipTiered)
	localID := c.cfg.LocalProviderID()

	if useLocal {
		if ep, ok := c.endpoints[localID]; ok {
			meta := PromptMeta{Provider: localID, Model: ep.model, Mode: aiModeLocalFirst}
			reply, thinking, err := c.completeAgent(ctx, ep, build(meta))
			if err == nil {
				return reply, thinking, nil
			}
			if order := router.ResolveTryOrder(c.cfg, c.sessionProvider, cloudKeyedSet(c.endpoints)); len(order) > 0 {
				router.LogFailover(nil, localID, order[0], err)
			}
		}
	}

	return c.chatWithFailoverAgent(ctx, build)
}

func (c *Client) chatWithFailoverAgent(ctx context.Context, build func(PromptMeta) []openai.ChatCompletionMessage) (string, string, error) {
	order := router.ResolveTryOrder(c.cfg, c.sessionProvider, cloudKeyedSet(c.endpoints))
	if len(order) == 0 {
		return "", "", fmt.Errorf("%s", errNoCloudProvidersWithKeys)
	}

	var errs []error
	var lastID ProviderID
	for i, id := range order {
		lastID = id
		ep := c.endpoints[id]
		meta := PromptMeta{Provider: id, Model: ep.model, Mode: aiModeCloudOnly}
		reply, thinking, err := c.completeAgent(ctx, ep, build(meta))
		if err == nil {
			return reply, thinking, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", id, err))
		if i+1 < len(order) {
			router.LogFailover(nil, id, order[i+1], err)
		}
	}
	return "", "", router.AggregateErrors(lastID, errs)
}

func (c *Client) chatWithFailover(ctx context.Context, build func(PromptMeta) []openai.ChatCompletionMessage) (string, string, error) {
	order := router.ResolveTryOrder(c.cfg, c.sessionProvider, cloudKeyedSet(c.endpoints))
	if len(order) == 0 {
		return "", "", fmt.Errorf("%s", errNoCloudProvidersWithKeys)
	}

	var errs []error
	var lastID ProviderID
	for i, id := range order {
		lastID = id
		ep := c.endpoints[id]
		meta := PromptMeta{Provider: id, Model: ep.model, Mode: aiModeCloudOnly}
		reply, thinking, err := c.complete(ctx, ep, build(meta))
		if err == nil {
			return reply, thinking, nil
		}
		errs = append(errs, fmt.Errorf("%s: %w", id, err))
		if i+1 < len(order) {
			router.LogFailover(nil, id, order[i+1], err)
		}
	}
	return "", "", router.AggregateErrors(lastID, errs)
}

func (c *Client) complete(ctx context.Context, ep *endpoint, messages []openai.ChatCompletionMessage) (string, string, error) {
	if ep.apiStyle == config.APIStyleAnthropic {
		return ep.chatAnthropic(ctx, messages, false)
	}
	return ep.chatOpenAI(ctx, messages)
}

func (c *Client) completeAgent(ctx context.Context, ep *endpoint, messages []openai.ChatCompletionMessage) (string, string, error) {
	if ep.apiStyle == config.APIStyleAnthropic {
		return ep.chatAnthropic(ctx, messages, true)
	}
	return ep.chatOpenAIAgent(ctx, messages)
}

// Assist sends userPrompt with optional scrollback and remoteCwd context via cloud providers only.
func (c *Client) Assist(ctx context.Context, userPrompt, scrollback, remoteCwd string) (string, error) {
	reply, _, err := c.chatWithFailover(ctx, func(meta PromptMeta) []openai.ChatCompletionMessage {
		return assistMessages(userPrompt, scrollback, remoteCwd, meta)
	})
	return reply, err
}

// SuggestInstall asks the AI for brief install instructions for command on the local OS (cloud only).
func (c *Client) SuggestInstall(ctx context.Context, command string) (string, error) {
	prompt := fmt.Sprintf(
		"The command %q is not installed on this system (%s). "+
			"Give brief install instructions for the most common package manager on this OS. "+
			"Do not run commands; only suggest.",
		command, platform.OSDescription(),
	)
	return c.Assist(ctx, prompt, "", "")
}

// StatusLine returns a one-line summary of the active AI routing for REPL model command.
func (c *Client) StatusLine() string {
	meta := c.promptMetaForStatus()
	override := ""
	if c.sessionProvider != "" {
		override = fmt.Sprintf(statusSessionOverride, c.sessionProvider)
	}
	tiered := ""
	if c.TieredEnabled() && c.sessionProvider == "" {
		tiered = statusTieredOnSuffix
	}
	return fmt.Sprintf(statusAIProviderPrefix+"%s, model: %s, mode: %s (config default: %s%s%s)",
		meta.Provider, meta.Model, meta.Mode, c.cfg.DefaultProviderID(), override, tiered)
}

func (c *Client) promptMetaForStatus() PromptMeta {
	if c.sessionProvider != "" {
		ep := c.endpoints[c.sessionProvider]
		return PromptMeta{Provider: c.sessionProvider, Model: ep.model, Mode: aiModeCloudOnly}
	}
	if c.TieredEnabled() {
		localID := c.cfg.LocalProviderID()
		ep := c.endpoints[localID]
		return PromptMeta{Provider: localID, Model: ep.model, Mode: aiModeTieredLabel}
	}
	id := c.cfg.DefaultProviderID()
	ep := c.endpoints[id]
	return PromptMeta{Provider: id, Model: ep.model, Mode: aiModeCloudOnly}
}
