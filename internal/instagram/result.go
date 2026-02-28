package instagram

import (
	"encoding/json"
	"fmt"
	"os"
)

func WriteResult(comments []Comment) error {
	if len(comments) > 0 {
		_ = os.MkdirAll("../../result", 0755)
		f, err := os.Create("../../result/result.json")
		if err == nil {
			defer f.Close()
			enc := json.NewEncoder(f)
			enc.SetIndent("", "  ")
			if err := enc.Encode(comments); err != nil {
				return fmt.Errorf("failed to encode comments: %w", err)
			}
			return nil
		} else {
			return fmt.Errorf("failed to create result file: %w", err)
		}
	}
	return nil
}
