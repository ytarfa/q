package main

// truncateResponse applies the hard character limit to the response.
// If limit is 0, no truncation is performed.
// Truncation happens at the last word boundary within the limit.
// If no word boundary is found, the response is hard-cut at the limit.
// An ellipsis "..." is appended to truncated responses (not counted against the limit).
func truncateResponse(response string, limit int) string {
	if limit <= 0 || len(response) <= limit {
		return response
	}

	// Find the last space within the limit
	truncated := response[:limit]
	lastSpace := -1
	for i := len(truncated) - 1; i >= 0; i-- {
		if truncated[i] == ' ' {
			lastSpace = i
			break
		}
	}

	if lastSpace > 0 {
		return truncated[:lastSpace] + "..."
	}

	// No word boundary found -- hard-cut at limit
	return truncated + "..."
}
