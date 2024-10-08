package interpolatetext

import (
	"errors"
	"reflect"
	"testing"
)

type mockInterpolateApplyCallableN struct {
	arg string
}

func (m *mockInterpolateApplyCallableN) apply(data interface{}) (result string, err error) {
	return m.arg, nil
}

func mockInterpolateArgParserN(arg string) (callable interpolateApplyCallable, err error) {
	callable = &mockInterpolateApplyCallableN{
		arg: arg,
	}
	return callable, nil
}

func checkMockInterpolateApplyCallableN(t *testing.T, part interpolateApplyCallable, expectArg string) {
	var inst *mockInterpolateApplyCallableN
	var ok bool
	if inst, ok = part.(*mockInterpolateApplyCallableN); !ok {
		t.Fatalf("expect mock apply callable N [%s] but get wrong type apply callable: %v", expectArg, reflect.TypeOf(part))
	}
	if expectArg != inst.arg {
		t.Errorf("expect mock apply callable N [%s] but get different argument [%s]", expectArg, inst.arg)
	}
}

func checkLiteral(t *testing.T, part interpolateApplyCallable, expectText string) {
	var inst *literalInterpolateApply
	var ok bool
	if inst, ok = part.(*literalInterpolateApply); !ok {
		t.Fatalf("expect literal [%s] but get wrong type apply callable: %v", expectText, reflect.TypeOf(part))
	}
	s := *((*string)(inst))
	if expectText != s {
		t.Errorf("expect literal [%s] but get different content [%s]", expectText, s)
	}
}

type caseN struct {
	isLiteral     bool
	expectContent string
}

func newCaseN(isLiteral bool, expectContent string) (result *caseN) {
	return &caseN{
		isLiteral:     isLiteral,
		expectContent: expectContent,
	}
}

func runTestOfCaseN(t *testing.T, templateText string, testPlan []*caseN) {
	engine := newTemplateParseEngine(mockInterpolateArgParserN)
	if err := engine.parse(templateText); nil != err {
		t.Fatalf("caught error on parsing: %v", err)
	}
	parts := engine.interpolateParts
	if len(testPlan) != len(parts) {
		t.Fatalf("length of interpolate parts not correct: %d (expect %d)", len(parts), len(testPlan))
	}
	for idx, c := range testPlan {
		if c.isLiteral {
			checkLiteral(t, parts[idx], c.expectContent)
		} else {
			checkMockInterpolateApplyCallableN(t, parts[idx], c.expectContent)
		}
	}
}

func TestTemplateParseEngineCaseN0(t *testing.T) {
	runTestOfCaseN(t, "${ABCdEf}", []*caseN{
		newCaseN(false, "ABCdEf"),
	})
}

func TestTemplateParseEngineCaseN1(t *testing.T) {
	runTestOfCaseN(t, "${ABCdEf}123", []*caseN{
		newCaseN(false, "ABCdEf"),
		newCaseN(true, "123"),
	})
}

func TestTemplateParseEngineCaseN2(t *testing.T) {
	runTestOfCaseN(t, "Abc${dEf}123", []*caseN{
		newCaseN(true, "Abc"),
		newCaseN(false, "dEf"),
		newCaseN(true, "123"),
	})
}

func TestTemplateParseEngineCaseN3(t *testing.T) {
	runTestOfCaseN(t, "${dEf}123", []*caseN{
		newCaseN(false, "dEf"),
		newCaseN(true, "123"),
	})
}

func TestTemplateParseEngineCaseN4(t *testing.T) {
	runTestOfCaseN(t, "${dEf}123${Ghi}GK", []*caseN{
		newCaseN(false, "dEf"),
		newCaseN(true, "123"),
		newCaseN(false, "Ghi"),
		newCaseN(true, "GK"),
	})
}

func TestTemplateParseEngineCaseN5(t *testing.T) {
	runTestOfCaseN(t, "${dEf}123${Ghi}GK\\$ABC\\${defghi}", []*caseN{
		newCaseN(false, "dEf"),
		newCaseN(true, "123"),
		newCaseN(false, "Ghi"),
		newCaseN(true, "GK$ABC${defghi}"),
	})
}

func TestTemplateParseEngineCaseN6(t *testing.T) {
	runTestOfCaseN(t, "{dEf}123{Ghi}GK\\$ABC", []*caseN{
		newCaseN(true, "{dEf}123{Ghi}GK$ABC"),
	})
}

func TestTemplateParseEngineCaseN7(t *testing.T) {
	runTestOfCaseN(t, "${dEf}${Ghi}GKrrr", []*caseN{
		newCaseN(false, "dEf"),
		newCaseN(false, "Ghi"),
		newCaseN(true, "GKrrr"),
	})
}

type mockInterpolateApplyCallableP struct {
	arg string
}

func (m *mockInterpolateApplyCallableP) apply(data interface{}) (result string, err error) {
	if nil == data {
		return "[" + m.arg + "]", nil
	}
	if err, ok := data.(error); ok {
		return m.arg, err
	}
	if textMap, ok := data.(map[string]string); ok {
		return "(" + textMap[m.arg] + ")", nil
	}
	return "/" + m.arg + "/", nil
}

func mockInterpolateArgParserP(arg string) (callable interpolateApplyCallable, err error) {
	callable = &mockInterpolateApplyCallableP{
		arg: arg,
	}
	return callable, nil
}

func checkMockInterpolateApplyCallablePNoErrorTemplate(t *testing.T, templateText string, data interface{}, expectText string) {
	var tpl templateBase
	if err := tpl.parseTemplate(templateText, mockInterpolateArgParserP); nil != err {
		t.Fatalf("failed on parsing template [%s]: %v", templateText, err)
	}
	result, err := tpl.applyContent(data, false)
	if result != expectText {
		t.Errorf("result content not expect [%s] != [%s]", result, expectText)
	}
	if nil != err {
		t.Errorf("not expecting error but having error: %v", err)
	}
}

func TestTemplateBaseCaseP0(t *testing.T) {
	checkMockInterpolateApplyCallablePNoErrorTemplate(t, "Abc${DEF}Ghi", nil, "Abc[DEF]Ghi")
	checkMockInterpolateApplyCallablePNoErrorTemplate(t, "Abc${DEF}Ghi", errors.New("test error"), "Abc${DEF}Ghi")
	var tmap = map[string]string{
		"DEF": "apple",
		"ABC": "banana",
	}
	checkMockInterpolateApplyCallablePNoErrorTemplate(t, "Abc${DEF}Ghi", tmap, "Abc(apple)Ghi")
	checkMockInterpolateApplyCallablePNoErrorTemplate(t, "Abc${DEF}Ghi\\$${ABC}", tmap, "Abc(apple)Ghi$(banana)")
	var emap = map[string]string{
		"ABC": "banana",
	}
	checkMockInterpolateApplyCallablePNoErrorTemplate(t, "Abc${DEF}Ghi", emap, "Abc()Ghi")
	checkMockInterpolateApplyCallablePNoErrorTemplate(t, "Abc${DEF}Ghi", "else-case", "Abc/DEF/Ghi")
}

func checkMockInterpolateApplyCallablePNoErrorApplyContent(t *testing.T, templateText string, data interface{}, expectText string) {
	result, err := applyContent(templateText, data, mockInterpolateArgParserP, false)
	if result != expectText {
		t.Errorf("result content not expect [%s] != [%s]", result, expectText)
	}
	if nil != err {
		t.Errorf("not expecting error but having error: %v", err)
	}
}

func TestApplyContentCaseP0(t *testing.T) {
	checkMockInterpolateApplyCallablePNoErrorApplyContent(t, "Abc${DEF}Ghi", nil, "Abc[DEF]Ghi")
	checkMockInterpolateApplyCallablePNoErrorApplyContent(t, "Abc${DEF}Ghi", errors.New("test error"), "Abc${DEF}Ghi")
	var tmap = map[string]string{
		"DEF": "apple",
		"ABC": "banana",
	}
	checkMockInterpolateApplyCallablePNoErrorApplyContent(t, "Abc${DEF}Ghi", tmap, "Abc(apple)Ghi")
	checkMockInterpolateApplyCallablePNoErrorApplyContent(t, "Abc${DEF}Ghi\\$${ABC}", tmap, "Abc(apple)Ghi$(banana)")
	var emap = map[string]string{
		"ABC": "banana",
	}
	checkMockInterpolateApplyCallablePNoErrorApplyContent(t, "Abc${DEF}Ghi", emap, "Abc()Ghi")
	checkMockInterpolateApplyCallablePNoErrorApplyContent(t, "Abc${DEF}Ghi", "else-case", "Abc/DEF/Ghi")
}
