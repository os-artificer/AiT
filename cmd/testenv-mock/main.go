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

// testenv-mock serves loopback SSH+SFTP and HTTP mocks for automated acceptance tests.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/os-artificer/AiT/internal/pkg/testhttp"
	"github.com/os-artificer/AiT/internal/pkg/testssh"
)

func main() {
	sshAddr := flag.String("ssh-addr", testssh.DefaultAddr, "SSH listen address")
	httpAddr := flag.String("http-addr", "127.0.0.1:0", "HTTP listen address")
	root := flag.String("root", "", "SSH filesystem root (default: temp dir)")
	user := flag.String("user", testssh.DefaultUser, "SSH username")
	password := flag.String("password", testssh.DefaultPassword, "SSH password")
	reject := flag.Bool("reject-password", false, "reject all SSH passwords")
	scenario := flag.String("http-scenario", testhttp.ScenarioOK, "HTTP scenario (ok, unauthorized, rate_limited, server_error, not_found, malformed, geocode_unable, hang)")
	flag.Parse()

	rootDir := *root
	if rootDir == "" {
		dir, err := os.MkdirTemp("", "ait-testssh-*")
		if err != nil {
			log.Fatal(err)
		}
		rootDir = dir
	}
	idDir, err := os.MkdirTemp("", "ait-testssh-id-*")
	if err != nil {
		log.Fatal(err)
	}

	sshSrv, err := testssh.Listen(testssh.Config{
		Addr:           *sshAddr,
		User:           *user,
		Password:       *password,
		RejectPassword: *reject,
		RootDir:        rootDir,
		IdentityDir:    idDir,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sshSrv.Close()

	httpSrv, err := testhttp.Listen(testhttp.Config{
		Addr:     *httpAddr,
		Scenario: *scenario,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer httpSrv.Close()

	payload := map[string]any{
		"ssh":  sshSrv.Info(),
		"http": httpSrv.Info(),
	}
	enc, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(enc))
	_ = os.Stdout.Sync()
	fmt.Fprintf(os.Stderr, "testenv-mock ssh %s http %s root %s\n", sshSrv.Addr(), httpSrv.BaseURL(), filepath.Clean(rootDir))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
