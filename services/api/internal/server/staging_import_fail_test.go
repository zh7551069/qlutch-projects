package server

import (
	"context"
	"net/http"
	"strings"
	"testing"

	qlutchapi "github.com/zh7551069/qlutch-projects/packages/contracts/gen/go"
)

func TestStagingImportFailIsRepeatableAndLeavesTheMainPath(t *testing.T) {
	retryable := stageMaterial(t, "13800018101", true, "paste_note", "text/plain", []byte("无法完成挖掘"))
	assertFailed(t, retryable, qlutchapi.JobStatusFailedRetryable, true, 0)
	padded := stageMaterial(t, "13800018102", true, "paste_note", "text/plain", []byte("  无法完成挖掘\n"))
	assertFailed(t, padded, qlutchapi.JobStatusFailedRetryable, true, 0)
	terminal := stageMaterial(t, "13800018103", true, "paste_note", "text/plain", []byte("材料无法读取"))
	assertFailed(t, terminal, qlutchapi.JobStatusFailedTerminal, false, 0)

	ordinary := stageMaterial(t, "13800018104", true, "paste_note", "text/plain", []byte("叶绿体把光能写成糖"))
	assertReady(t, ordinary)
	near := stageMaterial(t, "13800018105", true, "paste_note", "text/plain", []byte("无法完成挖掘。"))
	assertReady(t, near)
	photo := stageMaterial(t, "13800018106", true, "camera", "image/jpeg", []byte("无法完成挖掘"))
	assertReady(t, photo)

	off := stageMaterial(t, "13800018107", false, "paste_note", "text/plain", []byte("无法完成挖掘"))
	assertReady(t, off)
	offTerminal := stageMaterial(t, "13800018108", false, "paste_note", "text/plain", []byte("材料无法读取"))
	assertReady(t, offTerminal)
}

type stagedMaterial struct {
	job   qlutchapi.APIResponse[qlutchapi.KnowledgeJob]
	s01   int
	enter bool
}

func stageMaterial(t *testing.T, phone string, staging bool, source, contentType string, body []byte) stagedMaterial {
	t.Helper()
	h := newKnowledge(t)
	if staging {
		h.know.SetStagingImportFail(true)
	}
	access, _, _ := h.register(t, phone)
	h.onboard(t, access)
	pack := h.createPack(t, access).Body.Pack_id
	upload := h.upload(t, access, pack, source, contentType, int64(len(body)), "")
	key := strings.TrimPrefix(upload.Body.Upload_url, "http://127.0.0.1/upload/")
	h.blobs.Put(key, body)
	imported := h.importJob(t, access, "idem-"+phone, pack, upload.Body.Material_id)
	if imported.StatusCode != http.StatusAccepted || imported.Body.Map_ready || imported.Body.Account.Has_first_pack {
		t.Fatal("import acceptance changed the main path")
	}
	provider := &flowProvider{}
	attachRuntime(h, provider)
	if err := h.know.ProcessQueued(context.Background(), h.queue.Items()[0]); err != nil {
		t.Fatal(err)
	}
	got := h.job(t, access, imported.Body.Job_id)
	enter, err := h.know.TestingMapEnterable(context.Background(), pack)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got.Raw), "sk-") {
		t.Fatal("job response contained a fixture secret marker")
	}
	return stagedMaterial{job: got, s01: provider.s01, enter: enter}
}

func assertFailed(t *testing.T, got stagedMaterial, status qlutchapi.JobStatus, retryable bool, calls int) {
	t.Helper()
	body := got.job.Body
	if body.Status != status || body.Retryable != retryable || body.Map_ready || body.Account.Has_first_pack || got.s01 != calls || got.enter {
		t.Fatalf("status=%s retry=%v map=%v pack=%v calls=%d enter=%v", body.Status, body.Retryable, body.Map_ready, body.Account.Has_first_pack, got.s01, got.enter)
	}
}

func assertReady(t *testing.T, got stagedMaterial) {
	t.Helper()
	body := got.job.Body
	if body.Status != qlutchapi.JobStatusMapReady || !body.Map_ready || body.Retryable || !body.Account.Has_first_pack || got.s01 != 1 || !got.enter {
		t.Fatalf("status=%s retry=%v map=%v pack=%v calls=%d enter=%v", body.Status, body.Retryable, body.Map_ready, body.Account.Has_first_pack, got.s01, got.enter)
	}
}
