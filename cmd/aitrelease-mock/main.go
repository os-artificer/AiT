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

// aitrelease-mock serves a fake GitHub releases/latest API for local sandbox testing.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/aitrelease"
	"github.com/os-artificer/AiT/internal/pkg/aitrelease/testrelease"
)

func main() {
	addr := flag.String("addr", ":8765", "listen address")
	tag := flag.String("tag", "v0.9.0-sandbox", "release tag_name")
	archive := flag.String("archive", "", "optional release archive path (tar.gz or zip); default stub bundle")
	flag.Parse()

	baseURL, shutdown, err := testrelease.ListenAndServe(*addr, testrelease.ServeOptions{
		Tag:         *tag,
		ArchivePath: *archive,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("mock GitHub API listening on %s\n", baseURL)
	fmt.Printf("export %s=%s\n", aitrelease.TestReleaseBaseURLEnv, baseURL)
	fmt.Println("Ctrl+C to stop")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
