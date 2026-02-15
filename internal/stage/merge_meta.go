package stage

import "context"

const mergeMetaStage = "merge-meta"

func mergeMetaRunner(ctx context.Context, in Envelope, deps Deps) (Envelope, error) {
	out := in
	for i, r := range in.Records {
		if r.Error != nil {
			continue
		}
		var next map[string]any
		if r.Post != nil {
			if pm, ok := r.Post.(map[string]any); ok {
				if em, ok2 := pm["existingMeta"].(map[string]any); ok2 {
					next = em
				}
			}
		}
		if next == nil {
			next = map[string]any{}
		}
		m := map[string]any{"nextMeta": next}
		if r.Post != nil {
			if pm, ok := r.Post.(map[string]any); ok {
				for k, v := range pm {
					m[k] = v
				}
			}
		}
		r.Post = m
		out.Records[i] = r
	}
	return out, nil
}

func init() { Register(mergeMetaStage, mergeMetaRunner) }
