package interpolatetext

import (
	"errors"
	"fmt"
	"strings"
)

// ErrTextMapInterpolateKeyNotFound raises when given interpolate key cannot be
// mapped in given text map.
type ErrTextMapInterpolateKeyNotFound string

func (err ErrTextMapInterpolateKeyNotFound) Error() string {
	return fmt.Sprintf("cannot found key [%v] in given text map", (string)(err))
}

// ErrCannotConvertDataIntoTextMap raises when given interpolate data cannot be
// convert into text map.
var ErrCannotConvertDataIntoTextMap = errors.New("cannot convert given data into text map")

// TextMapInterpolation is a instance of text-mapping based interpolation
// template.
type TextMapInterpolation struct {
	templateBase
}

type textMapInterpolationKey string

func (k textMapInterpolationKey) apply(data interface{}) (result string, err error) {
	t := (string)(k)
	if m, ok := data.(map[string]string); ok {
		if result, ok = m[t]; ok {
			return result, nil
		}
		return t, (ErrTextMapInterpolateKeyNotFound)(t)
	}
	return t, ErrCannotConvertDataIntoTextMap
}

func textMapInterploateArgumentParser(arg string) (callable interpolateApplyCallable, err error) {
	arg = strings.ToUpper(arg)
	return (textMapInterpolationKey)(arg), nil
}

// NewTextMapInterpolation creates an instance of TextMapInterpolation with
// given template text.
func NewTextMapInterpolation(templateText string) (tpl *TextMapInterpolation, err error) {
	tpl = &TextMapInterpolation{}
	if err = tpl.parseTemplate(templateText, textMapInterploateArgumentParser); nil != err {
		return nil, err
	}
	return tpl, nil
}

// Apply interpolates given text map with template instance. Key of given text
// map must in upper-case.
func (tpl *TextMapInterpolation) Apply(textMap map[string]string, raiseError bool) (result string, err error) {
	return tpl.applyContent(textMap, raiseError)
}

// TextMapInterpolationSlice is a slice of TextMapInterpolation to ease
// interpolation on list of strings.
type TextMapInterpolationSlice []*TextMapInterpolation

// NewTextMapInterpolationSlice creates an instance of TextMapInterpolationSlice
func NewTextMapInterpolationSlice(templateTexts []string) (tpl TextMapInterpolationSlice, err error) {
	result := make([]*TextMapInterpolation, 0, len(templateTexts))
	for _, templateText := range templateTexts {
		t, err := NewTextMapInterpolation(templateText)
		if nil != err {
			return nil, err
		}
		result = append(result, t)
	}
	return result, nil
}

// Apply interpolates given text map with template instance. Key of given text
// map must in upper-case.
func (tpl TextMapInterpolationSlice) Apply(textMap map[string]string, raiseError bool) (result []string, err error) {
	result = make([]string, 0, len(tpl))
	for _, t := range tpl {
		var r string
		if r, err = t.Apply(textMap, raiseError); nil != err {
			return nil, err
		}
		result = append(result, r)
	}
	return result, nil
}
