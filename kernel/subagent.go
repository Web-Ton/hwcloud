package kernel

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	hwcloud "github.com/Cloud-Developer-Department/hwcloud"
	"github.com/Cloud-Developer-Department/hwcloud/agent"
)

// globalAgentSeq generates unique sub-agent session ids across runtimes.
var globalAgentSeq atomic.Int64

// runChild runs a nested Runtime as a sub-agent: fresh or stable session id
// under the parent's user/context scope, capped turns, no nested delegation.
// deps must be pre-built by the caller (tools resolved; MemoryProvider,
// policy, and approver inherited from the parent — v2.0 §22). emit
// receives child stream events for UI progress (nil = synchronous run).
// childSessionID, when non-empty, is reused across calls so the child's
// SessionStore history carries over (resumable sub-agent); empty generates
// a fresh one-shot id. The returned string is the child's final answer.
func runChild(ctx context.Context, cfg *agent.Agent, deps Deps, session hwcloud.Session, task string, emit func(hwcloud.StreamEvent), childSessionID string) (string, error) {
	child := session
	if childSessionID != "" {
		child.ID = childSessionID // stable: history persists in the child's store
	} else {
		child.ID = fmt.Sprintf("%s-%d", cfg.Name, globalAgentSeq.Add(1)) // one-shot
	}
	sub := New(cfg, deps)

	if emit == nil {
		res, err := sub.Run(ctx, child, hwcloud.UserMessage(task))
		if err != nil {
			return "", err
		}
		return res.FinalOutput, nil
	}

	var output strings.Builder
	subCh := sub.RunStream(ctx, child, hwcloud.UserMessage(task))
	for {
		select {
		case ev, ok := <-subCh:
			if !ok {
				// Channel closed without StreamDone: the run ended
				// abnormally — don't hand back a partial transcript as
				// success.
				return output.String(), fmt.Errorf("sub-agent %s: stream ended without a result", cfg.Name)
			}
			switch ev.Type {
			case hwcloud.StreamThought, hwcloud.StreamTextDelta:
				output.WriteString(ev.Text)
				emit(ev)
			case hwcloud.StreamToolResult:
				output.WriteString(ev.Message.Content)
				emit(ev)
			case hwcloud.StreamError:
				if ev.Error != nil {
					return output.String(), ev.Error
				}
			case hwcloud.StreamAborted:
				if ev.Error != nil {
					return output.String(), ev.Error
				}
				return output.String(), ctx.Err()
			case hwcloud.StreamDone:
				if ev.Result != nil {
					return ev.Result.FinalOutput, nil
				}
				return output.String(), nil
			}
		case <-ctx.Done():
			// Cancellation is an error, not a successful partial answer.
			return output.String(), ctx.Err()
		}
	}
}
