package interpolatetext

import "fmt"
import "errors"

// ErrEmptyInterpolateArgument represents error discovered on parsing template
// string where an empty interpolate (eg `${}`) is given.
type ErrEmptyInterpolateArgument struct {
	Position int
}

func newErrEmptyInterpolateArgument(position int) (err error) {
	return &ErrEmptyInterpolateArgument{
		Position: position,
	}
}

func (err *ErrEmptyInterpolateArgument) Error() string {
	return fmt.Sprintf("empty interpolate argument (position=%d)", err.Position)
}

// ErrInterpolateArgumentParseFailed represents error occurs in parsing
// interpolate argument.
type ErrInterpolateArgumentParseFailed struct {
	Position    int
	ParserError error
}

func newErrInterpolateArgumentParseFailed(position int, parserError error) (err error) {
	return &ErrInterpolateArgumentParseFailed{
		Position:    position,
		ParserError: parserError,
	}
}

func (err *ErrInterpolateArgumentParseFailed) Error() string {
	return fmt.Sprintf("interpolate argument parsing failed (position=%d, error=%v)", err.Position, err.ParserError)
}

var ErrBraceNotClose error = errors.New("brace in template is not close")

type interpolateApplyCallable interface {
	apply(data interface{}) string
}

type interpolateArgumentParser func(string) (interpolateApplyCallable, error)

type literalInterpolateApply string

func (literal *literalInterpolateApply) apply(interface{}) string {
	return (string)(*literal)
}

type templateParseEngine struct {
	interpolateParts []interpolateApplyCallable
	state            int
	partStart        int
	partFinish       int
	argumentParser   interpolateArgumentParser
}

func newTemplateParseEngine(argumentParser interpolateArgumentParser) (engine *templateParseEngine) {
	return &templateParseEngine{
		interpolateParts: make([]interpolateApplyCallable, 0),
		state:            parseStateInit,
		partStart:        0,
		partFinish:       0,
		argumentParser:   argumentParser,
	}
}

func (engine *templateParseEngine) restartPartTracking(idx int) {
	pos := idx + 1
	engine.partStart = pos
	engine.partFinish = pos
}

func (engine *templateParseEngine) getExtendedLiteralString(templateText string) string {
	l := templateText[engine.partStart:engine.partFinish]
	i := len(engine.interpolateParts) - 1
	if i < 0 {
		return l
	}
	if fnt, ok := engine.interpolateParts[i].(*literalInterpolateApply); ok {
		l = (string)(*fnt) + l
		engine.interpolateParts = engine.interpolateParts[:i]
	}
	return l
}

func (engine *templateParseEngine) extendLiteral(templateText string) {
	if engine.partStart == engine.partFinish {
		return
	}
	literalStr := engine.getExtendedLiteralString(templateText)
	literalPart := (literalInterpolateApply)(literalStr)
	engine.interpolateParts = append(engine.interpolateParts, &literalPart)
}

func (engine *templateParseEngine) onStateInit(idx int, ch rune, templateText string) {
	switch ch {
	case '$':
		if engine.partStart != idx {
			engine.partFinish = idx
		}
		engine.state = parseStateDollarSign
	case '\\':
		engine.partFinish = idx
		engine.extendLiteral(templateText)
		engine.restartPartTracking(idx)
		engine.state = parseStateBackSlash
	}
}

func (engine *templateParseEngine) onStateDollarSign(idx int, ch rune, templateText string) {
	switch ch {
	case '{':
		engine.extendLiteral(templateText)
		engine.restartPartTracking(idx)
		engine.state = parseStateBraceStarted
	default:
		engine.partFinish = engine.partStart
		engine.state = parseStateInit
	}
}

func (engine *templateParseEngine) onStateBraceStarted(idx int, ch rune, templateText string) (err error) {
	if '}' != ch {
		return nil
	}
	engine.partFinish = idx
	if engine.partStart == engine.partFinish {
		return newErrEmptyInterpolateArgument(idx)
	}
	argText := templateText[engine.partStart:engine.partFinish]
	var argObj interpolateApplyCallable
	if argObj, err = engine.argumentParser(argText); nil != err {
		return newErrInterpolateArgumentParseFailed(idx, err)
	}
	engine.interpolateParts = append(engine.interpolateParts, argObj)
	engine.restartPartTracking(idx)
	engine.state = parseStateInit
	return nil
}

func (engine *templateParseEngine) onStateBackSlash(idx int, templateText string) {
	engine.partFinish = idx
	engine.extendLiteral(templateText)
	engine.restartPartTracking(idx)
}

func (engine *templateParseEngine) parse(templateText string) (err error) {
	for idx, ch := range templateText {
		switch engine.state {
		case parseStateInit:
			engine.onStateInit(idx, ch, templateText)
		case parseStateDollarSign:
			engine.onStateDollarSign(idx, ch, templateText)
		case parseStateBraceStarted:
			if err = engine.onStateBraceStarted(idx, ch, templateText); nil != err {
				return err
			}
		case parseStateBackSlash:
			engine.state = parseStateInit
		}
	}
	if engine.state == parseStateBraceStarted {
		return ErrBraceNotClose
	}
	if l := len(templateText); engine.partStart < l {
		engine.partFinish = l
		engine.extendLiteral(templateText)
	}
	return nil
}

type templateBase struct {
	interpolateParts []interpolateApplyCallable
}

const (
	parseStateInit = iota
	parseStateDollarSign
	parseStateBraceStarted
	parseStateBackSlash
)

func (tpl *templateBase) parseTemplate(templateText string, argumentParser interpolateArgumentParser) (err error) {
	tplEngine := newTemplateParseEngine(argumentParser)
	if err = tplEngine.parse(templateText); nil != err {
		return err
	}
	tpl.interpolateParts = tplEngine.interpolateParts
	return nil
}
