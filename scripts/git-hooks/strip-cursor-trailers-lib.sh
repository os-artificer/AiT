# Shared helpers for stripping Cursor-injected git commit trailers.
# Sourced by prepare-commit-msg, commit-msg, and post-commit hooks.

strip_cursor_trailer_lines() {
	# Drop git trailer lines (Key: value) for Cursor attribution only.
	grep -v -E '^(Co-authored-by:.*([Cc]ursor|cursoragent@cursor\.com)|Made-with:.*[Cc]ursor)' || true
}

trim_trailing_blank_lines() {
	# Remove trailing empty lines (portable sed).
	sed -e :a -e '/^\n*$/{$d;N;ba' -e '}'
}

strip_cursor_trailers_from_file() {
	commit_msg_file=$1
	[ -n "$commit_msg_file" ] && [ -f "$commit_msg_file" ] || return 0

	tmp="${commit_msg_file}.cursor-strip"
	strip_cursor_trailer_lines < "$commit_msg_file" | trim_trailing_blank_lines > "$tmp" || true
	mv "$tmp" "$commit_msg_file"
}

message_has_cursor_trailer() {
	msg=$1
	echo "$msg" | grep -q -E '^(Co-authored-by:.*([Cc]ursor|cursoragent@cursor\.com)|Made-with:.*[Cc]ursor)'
}
