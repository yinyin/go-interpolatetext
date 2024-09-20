package interpolatetext

import (
	"testing"
)

func TestTextMapInterpolationD1(t *testing.T) {
	inst, err := NewTextMapInterpolation("Abc${dEf}Ghi${JK}L")
	if nil != err {
		t.Fatalf("failed on parsing template: %v", err)
	}
	var tmap = map[string]string{
		"DEF": "-www-",
		"JK":  "[JP-Skrt]",
	}
	result, err := inst.Apply(tmap, true)
	if nil != err {
		t.Fatalf("failed on apply text-map: %v", err)
	}
	expText := "Abc-www-Ghi[JP-Skrt]L"
	if result != expText {
		t.Fatalf("unexpect result [%v] vs. [%v]", result, expText)
	}
}

func TestTextMapInterpolationSliceD1(t *testing.T) {
	var tplText = []string{
		"Abc${dEf}Ghi${JK}L",
		"Mno${dEf}",
		"${JK}",
		"WWW",
	}
	inst, err := NewTextMapInterpolationSlice(tplText)
	if nil != err {
		t.Fatalf("failed on parsing template: %v", err)
	}
	var tmap = map[string]string{
		"DEF": "-www-",
		"JK":  "[JP-Skrt]",
	}
	result, err := inst.Apply(tmap, true)
	if nil != err {
		t.Fatalf("failed on apply text-map: %v", err)
	}
	var expTexts = []string{
		"Abc-www-Ghi[JP-Skrt]L",
		"Mno-www-",
		"[JP-Skrt]",
		"WWW",
	}
	for idx, r := range result {
		expText := expTexts[idx]
		if r != expText {
			t.Fatalf("unexpect result (index=%d) [%v] vs. [%v]", idx, r, expText)
		}
	}
}
