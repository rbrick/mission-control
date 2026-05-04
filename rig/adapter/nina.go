package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	nina "github.com/rbrick/mission-control/nina"
	"github.com/rbrick/mission-control/protocol"
)

type NINAAdapter struct {
	id       string
	host     string
	client   *nina.ClientWithResponses
	registry *Registry
}

func NewNINA(id, host string) (*NINAAdapter, error) {
	if id == "" {
		id = "nina"
	}
	if host == "" {
		return nil, fmt.Errorf("nina host is required")
	}
	client, err := nina.NewClientWithResponses(host)
	if err != nil {
		return nil, err
	}
	a := &NINAAdapter{id: id, host: host, client: client}
	a.registry = NewRegistry(
		Command{Namespace: "rig", Name: "get_status", Handler: a.getStatus},
		Command{Namespace: "mount", Name: "goto_radec", Handler: a.gotoRADec},
		Command{Namespace: "mount", Name: "park", Handler: a.park},
		Command{Namespace: "mount", Name: "unpark", Handler: a.unpark},
		Command{Namespace: "mount", Name: "abort", Handler: a.abort},
		Command{Namespace: "camera", Name: "capture", Handler: a.capture},
		Command{Namespace: "sequence", Name: "start", Handler: a.startSequence},
		Command{Namespace: "sequence", Name: "stop", Handler: a.stopSequence},
	)
	return a, nil
}

func (a *NINAAdapter) ID() string   { return a.id }
func (a *NINAAdapter) Type() string { return "nina" }
func (a *NINAAdapter) Capabilities() []protocol.Capability {
	return a.registry.Capabilities(a.id, a.Type())
}
func (a *NINAAdapter) Handle(ctx context.Context, namespace, command string, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, *protocol.Error) {
	return a.registry.Handle(ctx, namespace, command, data, progress)
}

func (a *NINAAdapter) Status(ctx context.Context) (map[string]interface{}, error) {
	status := map[string]interface{}{"connected": true, "adapter": "nina", "host": a.host}
	if resp, err := a.client.GetEquipmentCameraInfoWithResponse(ctx); err == nil {
		status["camera_status"] = resp.StatusCode()
	}
	if resp, err := a.client.GetEquipmentMountInfoWithResponse(ctx); err == nil {
		status["mount_status"] = resp.StatusCode()
	}
	if resp, err := a.client.GetSequenceStateWithResponse(ctx); err == nil {
		status["sequence_status"] = resp.StatusCode()
	}
	return status, nil
}

func (a *NINAAdapter) getStatus(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	return a.Status(ctx)
}

func (a *NINAAdapter) gotoRADec(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	var req struct {
		RAHours       *float64 `json:"ra_hours"`
		RADegrees     *float64 `json:"ra_degrees"`
		DecDegrees    float64  `json:"dec_degrees"`
		Center        *bool    `json:"center"`
		WaitForResult *bool    `json:"wait_for_result"`
	}
	_ = json.Unmarshal(data, &req)
	if req.RAHours == nil && req.RADegrees == nil {
		return nil, Fail("INVALID_ARGUMENT", "ra_hours or ra_degrees is required")
	}
	ra := 0.0
	if req.RADegrees != nil {
		ra = *req.RADegrees
	} else {
		ra = *req.RAHours * 15.0
	}
	wait := true
	if req.WaitForResult != nil {
		wait = *req.WaitForResult
	}
	center := false
	if req.Center != nil {
		center = *req.Center
	}
	resp, err := a.client.GetEquipmentMountSlewWithResponse(ctx, &nina.GetEquipmentMountSlewParams{Ra: ra, Dec: req.DecDegrees, WaitForResult: &wait, Center: &center})
	return responseResult("mount.goto_radec", resp, err)
}

func (a *NINAAdapter) park(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	resp, err := a.client.GetEquipmentMountParkWithResponse(ctx)
	return responseResult("mount.park", resp, err)
}
func (a *NINAAdapter) unpark(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	resp, err := a.client.GetEquipmentMountUnparkWithResponse(ctx)
	return responseResult("mount.unpark", resp, err)
}
func (a *NINAAdapter) abort(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	resp, err := a.client.GetEquipmentMountSlewStopWithResponse(ctx)
	return responseResult("mount.abort", resp, err)
}
func (a *NINAAdapter) capture(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	var req struct {
		Duration      *float64 `json:"duration"`
		Gain          *float32 `json:"gain"`
		Save          *bool    `json:"save"`
		Solve         *bool    `json:"solve"`
		WaitForResult *bool    `json:"wait_for_result"`
		ImageType     *string  `json:"image_type"`
		TargetName    *string  `json:"target_name"`
	}
	_ = json.Unmarshal(data, &req)
	wait := true
	if req.WaitForResult != nil {
		wait = *req.WaitForResult
	}
	save := true
	if req.Save != nil {
		save = *req.Save
	}
	params := &nina.GetEquipmentCameraCaptureParams{Duration: req.Duration, Gain: req.Gain, Save: &save, Solve: req.Solve, WaitForResult: &wait, TargetName: req.TargetName}
	if req.ImageType != nil {
		t := nina.GetEquipmentCameraCaptureParamsImageType(*req.ImageType)
		params.ImageType = &t
	}
	resp, err := a.client.GetEquipmentCameraCaptureWithResponse(ctx, params)
	return responseResult("camera.capture", resp, err)
}
func (a *NINAAdapter) startSequence(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	skip := false
	resp, err := a.client.GetSequenceStartWithResponse(ctx, &nina.GetSequenceStartParams{SkipValidation: &skip})
	return responseResult("sequence.start", resp, err)
}
func (a *NINAAdapter) stopSequence(ctx context.Context, data json.RawMessage, progress ProgressFunc) (map[string]interface{}, error) {
	resp, err := a.client.GetSequenceStopWithResponse(ctx)
	return responseResult("sequence.stop", resp, err)
}

type statusCoder interface{ StatusCode() int }

func responseResult(operation string, resp statusCoder, err error) (map[string]interface{}, error) {
	if err != nil {
		return nil, Fail("UNAVAILABLE", err.Error())
	}
	if resp == nil {
		return nil, Fail("UNAVAILABLE", "nina returned no response")
	}
	status := resp.StatusCode()
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return nil, Fail("UNAVAILABLE", fmt.Sprintf("nina returned status %d", status))
	}
	return map[string]interface{}{"operation": operation, "status": status}, nil
}
