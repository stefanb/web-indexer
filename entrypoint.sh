#!/bin/bash

# Initialize command
cmd="web-indexer"

# Check each expected environment variable and append it to the command if set
[[ -n "$BASE_URL" ]] && cmd="$cmd --base-url \"$BASE_URL\""
[[ -n "$CONFIG" ]] && cmd="$cmd --config \"$CONFIG\""
[[ -n "$DATE_FORMAT" ]] && cmd="$cmd --date-format \"$DATE_FORMAT\""
[[ -n "$INDEX_FILE" ]] && cmd="$cmd --index-file \"$INDEX_FILE\""
[[ "$LINK_TO_INDEX" == "true" ]] && cmd="$cmd --link-to-index"
[[ -n "$LOG_LEVEL" ]] && cmd="$cmd --log-level \"$LOG_LEVEL\""
[[ "$MINIFY" == "true" ]] && cmd="$cmd --minify"
[[ "$RECURSIVE" == "true" ]] && cmd="$cmd --recursive"
[[ -n "$SKIP" ]] && cmd="$cmd --skip \"$SKIP\""
[[ -n "$SOURCE" ]] && cmd="$cmd --source \"$SOURCE\""
[[ -n "$TARGET" ]] && cmd="$cmd --target \"$TARGET\""
[[ -n "$TEMPLATE" ]] && cmd="$cmd --template \"$TEMPLATE\""
[[ -n "$TITLE" ]] && cmd="$cmd --title \"$TITLE\""

# Debug: Print the command to be executed
echo "Executing command: $cmd"

# Execute the constructed command
eval $cmd $@
