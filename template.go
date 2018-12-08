package interpolatetext

import "fmt"

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

type interpolateApplyCallable interface {
	apply(data interface{}) string
}

type interpolateArgumentParser func(string) (interpolateApplyCallable, error)

type literalInterpolateApply string

func (literal *literalInterpolateApply) apply(interface{}) string {
	return (string)(*literal)
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
	interpolateParts := make([]interpolateApplyCallable, 0)
	state := parseStateInit
	partStart := 0
	partFinish := 0
	for idx, ch := range templateText {
		switch state {
		case parseStateInit:
			switch ch {
			case '$':
				if partStart != idx {
					partFinish = idx
				}
				state = parseStateDollarSign
			case '\\':
				state = parseStateBackSlash
			}
		case parseStateDollarSign:
			switch ch {
			case '{':
				if partStart != partFinish {
					literal := (literalInterpolateApply)(templateText[partStart:partFinish])
					interpolateParts = append(interpolateParts, &literal)
				}
				partStart = idx + 1
				partFinish = partStart
				state = parseStateBraceStarted
			default:
				partFinish = partStart
				state = parseStateInit
			}
		case parseStateBraceStarted:
			switch ch {
			case '}':
				partFinish = idx
				if partStart == partFinish {
					return newErrEmptyInterpolateArgument(idx)
				}
				argText := templateText[partStart:partFinish]
				if argObj, err := argumentParser(argText); nil != err {
					return newErrInterpolateArgumentParseFailed(idx, err)
				} else {
					interpolateParts = append(interpolateParts, argObj)
				}
				partStart = idx + 1
				partFinish = partStart
				state = parseStateInit
			}
		case parseStateBackSlash:
			partFinish = idx
			if partStart != partFinish {
				literal := (literalInterpolateApply)(templateText[partStart:partFinish])
				interpolateParts = append(interpolateParts, &literal)
			}
			partStart = idx + 1
			partFinish = partStart
		}
	}
	if partStart < len(templateText) {
		literal := (literalInterpolateApply)(templateText[partStart:])
		interpolateParts = append(interpolateParts, &literal)
	}
	return nil
}
