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

package deps

// InstallFFmpegCommand returns a shell install command for ffmpeg (includes ffprobe when packaged together).
func InstallFFmpegCommand(target Target) string {
	switch target {
	case TargetDarwinBrew:
		return "brew install ffmpeg"
	case TargetLinuxApt:
		return "sudo apt-get update && sudo apt-get install -y ffmpeg"
	case TargetLinuxDnf:
		return "sudo dnf install -y ffmpeg"
	case TargetLinuxYum:
		return "sudo yum install -y ffmpeg"
	case TargetLinuxApk:
		return "sudo apk add --no-cache ffmpeg"
	case TargetLinuxPacman:
		return "sudo pacman -Sy --noconfirm ffmpeg"
	case TargetWindowsChoco:
		return "choco install ffmpeg -y"
	case TargetWindowsWinget:
		return "winget install --id Gyan.FFmpeg -e --accept-package-agreements --accept-source-agreements"
	default:
		return ""
	}
}

// InstallTesseractCommand returns a shell install command for tesseract (+ chi_sim when apt).
func InstallTesseractCommand(target Target) string {
	switch target {
	case TargetDarwinBrew:
		return "brew install tesseract tesseract-lang"
	case TargetLinuxApt:
		return "sudo apt-get update && sudo apt-get install -y tesseract-ocr tesseract-ocr-chi-sim"
	case TargetLinuxDnf:
		return "sudo dnf install -y tesseract tesseract-langpack-chi_sim"
	case TargetLinuxYum:
		return "sudo yum install -y tesseract"
	case TargetLinuxApk:
		return "sudo apk add --no-cache tesseract-ocr tesseract-ocr-data-chi_sim"
	case TargetLinuxPacman:
		return "sudo pacman -Sy --noconfirm tesseract tesseract-data-chi_sim"
	case TargetWindowsChoco:
		return "choco install tesseract -y"
	case TargetWindowsWinget:
		return "winget install --id UB-Mannheim.TesseractOCR -e --accept-package-agreements --accept-source-agreements"
	default:
		return ""
	}
}

// ManualFFmpegURL is the fallback documentation link when no package manager is available.
const ManualFFmpegURL = "https://ffmpeg.org/download.html"

// ManualTesseractURL is the fallback documentation link for tesseract.
const ManualTesseractURL = "https://tesseract-ocr.github.io/tessdoc/Installation.html"
