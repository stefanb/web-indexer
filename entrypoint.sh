#!/bin/bash

# Initialize command
cmd="s3-index-generator"

# Check each expected environment variable and append it to the command if set
[[ -n "$BUCKET" ]] && cmd="$cmd -bucket $BUCKET"
[[ -n "$PREFIX" ]] && cmd="$cmd -prefix $PREFIX"
[[ "$LINK_TO_INDEX" == "true" ]] && cmd="$cmd -link-to-index"
[[ "$RELATIVE_LINKS" == "true" ]] && cmd="$cmd -relative-links"
[[ -n "$TITLE" ]] && cmd="$cmd -title \"$TITLE\""
[[ -n "$DATE_FORMAT" ]] && cmd="$cmd -date-format \"$DATE_FORMAT\""
[[ -n "$URL" ]] && cmd="$cmd -url $URL"
[[ -n "$STAGING_DIR" ]] && cmd="$cmd -staging-dir $STAGING_DIR"
[[ -n "$TEMPLATE" ]] && cmd="$cmd -template $TEMPLATE"
[[ -n "$CONFIG" ]] && cmd="$cmd -config $CONFIG"
[[ "$UPLOAD" == "true" ]] && cmd="$cmd -upload"

# Debug: Print the command to be executed
echo "Executing command: $cmd"

# Execute the constructed command
eval $cmd
