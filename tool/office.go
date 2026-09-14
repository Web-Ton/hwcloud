package tool

import (
	"fmt"

	hwcloud "github.com/Cloud-Developer-Department/hwcloud"
)

// officeToolText wraps a success string into a ToolResult.
func officeToolText(text string) *hwcloud.ToolResult {
	return &hwcloud.ToolResult{Content: text}
}

// officeToolError wraps an error message into a failed ToolResult.
func officeToolError(toolName, text string) *hwcloud.ToolResult {
	return hwcloud.ErrorResult(fmt.Errorf("%s: %s", toolName, text), false, "")
}

// NewOfficeTools returns the office tools (Word + Excel + PPT). workDir is
// the workspace root used for relative-path resolution by all read and write
// tools — relative output paths land in the workspace, not ~/Documents, so
// the agent can read back what it just wrote.
func NewOfficeTools(workDir string) []hwcloud.Tool {
	return []hwcloud.Tool{
		&wordReadTool{workDir: workDir},
		&wordWriteTool{workDir: workDir},
		&excelReadTool{workDir: workDir},
		&excelWriteTool{workDir: workDir},
		&pptxReadTool{workDir: workDir},
		&pptxWriteTool{workDir: workDir},
		&pptxTemplateAnalyzeTool{},
		&pptxTemplateFillTool{workDir: workDir},
	}
}
