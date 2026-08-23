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

package toolname

import "testing"

func TestMetaPrefix_Golden(t *testing.T) {
	if MetaPrefix != "!ait" {
		t.Fatalf("MetaPrefix=%q want !ait", MetaPrefix)
	}
}

func TestMetaLine(t *testing.T) {
	if MetaLine("") != "!ait" {
		t.Fatalf("bare meta=%q", MetaLine(""))
	}
	if MetaLine("get /r ./l") != "!ait get /r ./l" {
		t.Fatalf("meta line=%q", MetaLine("get /r ./l"))
	}
}

func TestCLI(t *testing.T) {
	if CLI("") != "aissh" {
		t.Fatalf("CLI()=%q", CLI(""))
	}
	if CLI("session --list") != "aissh session --list" {
		t.Fatalf("CLI(sub)=%q", CLI("session --list"))
	}
}
