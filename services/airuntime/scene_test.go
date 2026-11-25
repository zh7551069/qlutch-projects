package airuntime

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

type scripted struct {
	calls     []ProviderRequest
	err       error
	tokens    int
	tree      MapTree
	questions []Question
}

func (s *scripted) Complete(_ context.Context, req ProviderRequest) (ProviderResponse, error) {
	copied := req
	copied.Fragment = append([]byte(nil), req.Fragment...)
	s.calls = append(s.calls, copied)
	if s.err != nil {
		return ProviderResponse{}, s.err
	}
	tree := s.tree
	if len(tree.Catalogs) == 0 {
		tree = sampleMap()
	}
	return ProviderResponse{Tokens: s.tokens, Map: tree, Questions: append([]Question(nil), s.questions...)}, nil
}

func sampleMap() MapTree {
	return MapTree{Catalogs: []Catalog{{
		Title: "目录",
		Chapters: []Chapter{{
			Title:  "章节",
			Points: []Point{{Title: "考点"}},
		}},
	}}}
}

func readyRuntime(provider Provider) (*Runtime, *MemoryRecorder) {
	rt := New()
	rec := NewMemoryRecorder()
	rt.SetRecorder(rec)
	if provider != nil {
		rt.SetProvider(provider)
	}
	return rt, rec
}

func TestCallStillRejectsAMissingSceneBeforeTheProvider(t *testing.T) {
	rt := New()
	if rt.Status() != "未配置提供方" {
		t.Fatalf("status = %q", rt.Status())
	}
	_, err := rt.Call(context.Background(), LedgerEntry{})
	if !errors.Is(err, ErrSceneCodeRequired) {
		t.Fatalf("empty scene = %v", err)
	}
	_, err = rt.Call(context.Background(), LedgerEntry{SceneCode: "scaffold"})
	if !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatalf("provider = %v", err)
	}
}

func TestS01SplitsTextAndDoesNotSendTheWholeNote(t *testing.T) {
	provider := &scripted{tokens: 4}
	rt, rec := readyRuntime(provider)
	body := []byte("ALPHA-PARAGRAPH\n\nBETA-PARAGRAPH")
	out, err := rt.RunS01(context.Background(), S01Input{
		ContentType: "text/plain",
		Body:        body,
		PackID:      "pack_1",
		JobID:       "job_1",
	})
	if err != nil || !mapEnterable(out.Map) || out.CacheHit {
		t.Fatalf("out=%+v err=%v", out, err)
	}
	if len(provider.calls) != 2 {
		t.Fatalf("calls = %d", len(provider.calls))
	}
	for _, call := range provider.calls {
		if call.SceneCode != SceneS01 || call.Subclass != SubclassText || call.Tier != TierLight || call.Media != mediaText {
			t.Fatalf("call = %+v", call)
		}
		if bytes.Equal(call.Fragment, body) || (bytes.Contains(call.Fragment, []byte("ALPHA")) && bytes.Contains(call.Fragment, []byte("BETA"))) {
			t.Fatalf("fragment = %q", call.Fragment)
		}
	}
	rows, err := rec.List(context.Background(), Query{SceneCode: SceneS01, Subclass: SubclassText})
	if err != nil || len(rows) != 1 || !rows[0].Success || rows[0].CacheHit || rows[0].Tokens != 8 || rows[0].Caller != CallerLearner || rows[0].RetryIndex != 0 || rows[0].Fingerprint == "" {
		t.Fatalf("rows = %+v %v", rows, err)
	}
	if strings.Contains(rows[0].Fingerprint, "ALPHA") || rows[0].TemplateVersion == "" {
		t.Fatalf("row = %+v", rows[0])
	}
	ocr, err := rec.List(context.Background(), Query{Subclass: SubclassOCR})
	if err != nil || len(ocr) != 0 {
		t.Fatalf("ocr = %+v", ocr)
	}
}

func TestS01RoutesImageAndAudioWithoutInventingTheOther(t *testing.T) {
	image := &scripted{tokens: 1}
	rt, _ := readyRuntime(image)
	_, err := rt.RunS01(context.Background(), S01Input{ContentType: "image/jpeg", Body: []byte{0xff, 0xd8, 0xff, 0x00}, PackID: "p", JobID: "j-img"})
	if err != nil || len(image.calls) != 1 || image.calls[0].Subclass != SubclassOCR || image.calls[0].Tier != TierRecognize {
		t.Fatalf("image calls=%d err=%v", len(image.calls), err)
	}
	audio := &scripted{tokens: 2}
	rt, rec := readyRuntime(audio)
	body := bytes.Repeat([]byte{0x01}, 32)
	_, err = rt.RunS01(context.Background(), S01Input{ContentType: "audio/mpeg", Body: body, PackID: "p", JobID: "j-aud"})
	if err != nil || len(audio.calls) != 1 || audio.calls[0].Subclass != SubclassASR || !bytes.Equal(audio.calls[0].Fragment, body) {
		t.Fatalf("audio err=%v calls=%d", err, len(audio.calls))
	}
	rows, err := rec.List(context.Background(), Query{SceneCode: SceneS01})
	if err != nil || len(rows) != 1 || rows[0].MetricKind != metricBytes || rows[0].MetricValue != len(body) || rows[0].Subclass != SubclassASR {
		t.Fatalf("audio row = %+v %v", rows, err)
	}
}

func TestS01SplitsADocument(t *testing.T) {
	provider := &scripted{tokens: 1}
	rt, _ := readyRuntime(provider)
	body := bytes.Repeat([]byte("a"), maxDocumentFragment*2+10)
	_, err := rt.RunS01(context.Background(), S01Input{ContentType: "application/pdf", Body: body, PackID: "p", JobID: "j-pdf"})
	if err != nil || len(provider.calls) != 3 {
		t.Fatalf("calls=%d err=%v", len(provider.calls), err)
	}
	for _, call := range provider.calls {
		if call.Subclass != SubclassText || call.Tier != TierStandard || bytes.Equal(call.Fragment, body) {
			t.Fatalf("call len=%d subclass=%s", len(call.Fragment), call.Subclass)
		}
	}
}

func TestS01FailureAndUnconfiguredProviderAreLedgered(t *testing.T) {
	rt, rec := readyRuntime(nil)
	_, err := rt.RunS01(context.Background(), S01Input{ContentType: "text/plain", Body: []byte("笔记"), PackID: "p", JobID: "j"})
	if !errors.Is(err, ErrProviderNotConfigured) {
		t.Fatalf("err = %v", err)
	}
	rows, err := rec.List(context.Background(), Query{JobID: "j"})
	if err != nil || len(rows) != 1 || rows[0].Success || rows[0].Tokens != 0 || rows[0].Subclass != SubclassText || rows[0].CacheHit {
		t.Fatalf("rows = %+v %v", rows, err)
	}

	provider := &scripted{err: errors.New("sk-secret 教材原文")}
	rt, rec = readyRuntime(provider)
	_, err = rt.RunS01(context.Background(), S01Input{ContentType: "text/plain", Body: []byte("笔记"), PackID: "p", JobID: "j2"})
	if err == nil {
		t.Fatal("provider failure was ignored")
	}
	rows, err = rec.List(context.Background(), Query{JobID: "j2"})
	if err != nil || len(rows) != 1 || rows[0].Success {
		t.Fatalf("failure row = %+v %v", rows, err)
	}
	if strings.Contains(rows[0].Fingerprint, "教材") || strings.Contains(rows[0].Subclass, "secret") {
		t.Fatalf("row leaked %s", rows[0].Fingerprint)
	}
}

func TestS01ReusesTheSameMaterialAndTemplate(t *testing.T) {
	provider := &scripted{tokens: 5}
	rt, rec := readyRuntime(provider)
	in := S01Input{ContentType: "text/plain", Body: []byte("同一份材料"), PackID: "p", JobID: "job-a"}
	if _, err := rt.RunS01(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	in.JobID = "job-b"
	out, err := rt.RunS01(context.Background(), in)
	if err != nil || !out.CacheHit || len(provider.calls) != 1 {
		t.Fatalf("reuse out=%+v calls=%d err=%v", out, len(provider.calls), err)
	}
	hit := true
	rows, err := rec.List(context.Background(), Query{SceneCode: SceneS01, CacheHit: true, CacheHitSet: hit})
	if err != nil || len(rows) != 1 || rows[0].JobID != "job-b" || rows[0].Tokens != 0 || !rows[0].Success {
		t.Fatalf("hit = %+v %v", rows, err)
	}
}

func TestS01RejectsAMapThatCopiesTheFragment(t *testing.T) {
	fragment := []byte("这一段很长，长到不该被当成地图标题存下来。")
	provider := &scripted{tree: MapTree{Catalogs: []Catalog{{
		Title:    string(fragment),
		Chapters: []Chapter{{Title: "章节", Points: []Point{{Title: "考点"}}}},
	}}}}
	rt, rec := readyRuntime(provider)
	_, err := rt.RunS01(context.Background(), S01Input{ContentType: "text/plain", Body: fragment, PackID: "p", JobID: "j"})
	if !errors.Is(err, ErrNotEnterable) {
		t.Fatalf("err = %v", err)
	}
	rows, err := rec.List(context.Background(), Query{JobID: "j"})
	if err != nil || len(rows) != 1 || rows[0].Success || strings.Contains(rows[0].Fingerprint, "这一段") {
		t.Fatalf("rows = %+v %v", rows, err)
	}
}

func TestS01NoTemplateOrStoppedSceneDoesNotPretend(t *testing.T) {
	provider := &scripted{tokens: 1}
	rt, rec := readyRuntime(provider)
	rt.setTemplate(SceneS01, "")
	_, err := rt.RunS01(context.Background(), S01Input{ContentType: "text/plain", Body: []byte("笔记"), PackID: "p", JobID: "j"})
	if !errors.Is(err, ErrNoTemplate) || len(provider.calls) != 0 {
		t.Fatalf("err=%v calls=%d", err, len(provider.calls))
	}
	rt.setTemplate(SceneS01, templateS01)
	rt.StopScene(SceneS01)
	_, err = rt.RunS01(context.Background(), S01Input{ContentType: "text/plain", Body: []byte("笔记"), PackID: "p", JobID: "j2"})
	if !errors.Is(err, ErrSceneStopped) || len(provider.calls) != 0 {
		t.Fatalf("stopped = %v calls=%d", err, len(provider.calls))
	}
	rows, err := rec.List(context.Background(), Query{SceneCode: SceneS01})
	if err != nil || len(rows) != 2 || rows[0].Success || rows[1].Success {
		t.Fatalf("rows = %+v %v", rows, err)
	}
}

func TestS01WithoutALedgerDoesNotCallTheProvider(t *testing.T) {
	provider := &scripted{}
	rt := New()
	rt.SetProvider(provider)
	_, err := rt.RunS01(context.Background(), S01Input{ContentType: "text/plain", Body: []byte("笔记"), PackID: "p", JobID: "j"})
	if !errors.Is(err, ErrLedger) || len(provider.calls) != 0 {
		t.Fatalf("err=%v calls=%d", err, len(provider.calls))
	}
}

func TestS02IsSeparateAndSmall(t *testing.T) {
	questions := []Question{{Stem: "一"}, {Stem: "二"}, {Stem: "三"}, {Stem: "四"}, {Stem: "五"}}
	provider := &scripted{tokens: 1, questions: questions}
	rt, rec := readyRuntime(provider)
	out, err := rt.RunS02(context.Background(), S02Input{
		Nodes:  []NodeRef{{ID: "n1", Title: "考点"}},
		PackID: "p",
		JobID:  "j",
	})
	if err != nil || len(out.Questions) != prefetchPerNode || len(provider.calls) != 1 {
		t.Fatalf("out=%d calls=%d err=%v", len(out.Questions), len(provider.calls), err)
	}
	if provider.calls[0].SceneCode != SceneS02 || provider.calls[0].Subclass != SubclassPrefetch || provider.calls[0].QuestionLimit != prefetchPerNode || provider.calls[0].QuestionLimit >= 100 || len(provider.calls[0].Fragment) != 0 {
		t.Fatalf("call = %+v", provider.calls[0])
	}
	rows, err := rec.List(context.Background(), Query{SceneCode: SceneS02, Subclass: SubclassPrefetch})
	if err != nil || len(rows) != 1 || !rows[0].Success || rows[0].SceneCode == SceneS01 {
		t.Fatalf("s02 = %+v %v", rows, err)
	}
	s01, err := rec.List(context.Background(), Query{SceneCode: SceneS01})
	if err != nil || len(s01) != 0 {
		t.Fatalf("s01 mixed = %+v", s01)
	}

	_, err = rt.RunS02(context.Background(), S02Input{
		Nodes:  []NodeRef{{ID: "a", Title: "1"}, {ID: "b", Title: "2"}, {ID: "c", Title: "3"}},
		PackID: "p",
		JobID:  "j-big",
	})
	if err == nil || len(provider.calls) != 1 {
		t.Fatalf("cap err=%v calls=%d", err, len(provider.calls))
	}
}

func TestLedgerRejectsMissingSceneSubclassAndOther(t *testing.T) {
	rt, rec := readyRuntime(nil)
	err := rt.append(context.Background(), Entry{SceneCode: "S05", Subclass: "讲题", Tier: TierLight, PackID: "p", JobID: "j", Caller: CallerLearner})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("s05 = %v", err)
	}
	err = rt.append(context.Background(), Entry{SceneCode: SceneS01, Subclass: "其他", Tier: TierLight, PackID: "p", JobID: "j", Caller: CallerLearner})
	if !errors.Is(err, ErrUsage) {
		t.Fatalf("other = %v", err)
	}
	rows, err := rec.List(context.Background(), Query{})
	if err != nil || len(rows) != 0 {
		t.Fatalf("rows = %+v", rows)
	}
	_, err = rt.RunS01(context.Background(), S01Input{ContentType: "video/mp4", Body: []byte("x"), PackID: "p", JobID: "j"})
	if !errors.Is(err, ErrSubclassRequired) {
		t.Fatalf("route = %v", err)
	}
}

func TestUsageShapeHasNoPricePromptOrSource(t *testing.T) {
	for _, typ := range []reflect.Type{reflect.TypeOf(Entry{}), reflect.TypeOf(ProviderRequest{}), reflect.TypeOf(LedgerEntry{})} {
		for i := 0; i < typ.NumField(); i++ {
			switch typ.Field(i).Name {
			case "Price", "UnitPrice", "Prompt", "Temperature", "Vendor", "Model", "APIKey", "Secret":
				t.Fatalf("%s has %s", typ.Name(), typ.Field(i).Name)
			}
		}
	}
}

func TestRetryIndexFollowsEarlierRowsForTheSameJob(t *testing.T) {
	provider := &scripted{err: errors.New("temporary")}
	rt, rec := readyRuntime(provider)
	in := S01Input{ContentType: "text/plain", Body: []byte("笔记"), PackID: "p", JobID: "job"}
	if _, err := rt.RunS01(context.Background(), in); err == nil {
		t.Fatal("expected failure")
	}
	provider.err = nil
	if _, err := rt.RunS01(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	rows, err := rec.List(context.Background(), Query{JobID: "job"})
	if err != nil || len(rows) != 2 || rows[0].RetryIndex != 0 || rows[1].RetryIndex != 1 || rows[0].Success || !rows[1].Success {
		t.Fatalf("rows = %+v %v", rows, err)
	}
}
