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
	"reflect"
	"testing"
)

func TestParseMetaChainLine_Login(t *testing.T) {
	chain, ok := ParseMetaChainLine("!ait login --save user@hostB")
	if !ok {
		t.Fatal("expected chain login")
	}
	if chain.Op != ChainOpLogin {
		t.Fatalf("op = %v, want ChainOpLogin", chain.Op)
	}
	want := []string{"--save", "user@hostB"}
	if !reflect.DeepEqual(chain.Args, want) {
		t.Fatalf("args = %v, want %v", chain.Args, want)
	}
}

func TestParseMetaChainLine_Load(t *testing.T) {
	chain, ok := ParseMetaChainLine("!ait load prod --tmux")
	if !ok {
		t.Fatal("expected chain load")
	}
	if chain.Op != ChainOpLoad {
		t.Fatalf("op = %v, want ChainOpLoad", chain.Op)
	}
	want := []string{"prod", "--tmux"}
	if !reflect.DeepEqual(chain.Args, want) {
		t.Fatalf("args = %v, want %v", chain.Args, want)
	}
}

func TestParseMetaChainLine_Switch(t *testing.T) {
	chain, ok := ParseMetaChainLine("!ait session --switch prod")
	if !ok {
		t.Fatal("expected chain switch")
	}
	if chain.Op != ChainOpSwitch {
		t.Fatalf("op = %v, want ChainOpSwitch", chain.Op)
	}
	if len(chain.Args) != 1 || chain.Args[0] != "prod" {
		t.Fatalf("args = %v", chain.Args)
	}
}

func TestParseMetaChainLine_NotChain(t *testing.T) {
	if _, ok := ParseMetaChainLine("!ait list"); ok {
		t.Fatal("list should not be chain action")
	}
	if _, ok := ParseMetaChainLine("!ait session --load prod"); ok {
		t.Fatal("session --load should not match switch chain")
	}
}
