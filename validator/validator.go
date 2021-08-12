package validator

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-courier/httptransport/validator/rules"
	typesutil "github.com/go-courier/x/types"
)

func MustParseRuleStringWithType(ruleStr string, typ typesutil.Type) *Rule {
	r, err := ParseRuleWithType([]byte(ruleStr), typ)
	if err != nil {
		panic(err)
	}
	return r
}

func ParseRuleWithType(ruleBytes []byte, typ typesutil.Type) (*Rule, error) {
	r := &rules.Rule{}

	if len(ruleBytes) != 0 {
		parsedRule, err := rules.ParseRule(ruleBytes)
		if err != nil {
			return nil, err
		}
		r = parsedRule
	}

	return &Rule{
		Type: typ,
		Rule: r,
	}, nil
}

type Rule struct {
	*rules.Rule

	ErrMsg []byte
	Type   typesutil.Type
}

func (r *Rule) SetOptional(optional bool) {
	r.Optional = optional
}

func (r *Rule) SetErrMsg(errMsg []byte) {
	r.ErrMsg = errMsg
}

func (r *Rule) SetDefaultValue(defaultValue []byte) {
	r.DefaultValue = defaultValue
}

func (r *Rule) String() string {
	return typesutil.FullTypeName(r.Type) + string(r.Rule.Bytes())
}

type RuleModifier interface {
	SetOptional(optional bool)
	SetDefaultValue(defaultValue []byte)
	SetErrMsg(errMsg []byte)
}

type RuleProcessor = func(rule RuleModifier)

// ValidatorMgr mgr for compiling validator
type ValidatorMgr interface {
	// Compile compile rule string to validator
	Compile(context.Context, []byte, typesutil.Type, ...RuleProcessor) (Validator, error)
}

var ValidatorMgrDefault = NewValidatorFactory()

type contextKeyValidatorMgr struct{}

func ContextWithValidatorMgr(c context.Context, validatorMgr ValidatorMgr) context.Context {
	return context.WithValue(c, contextKeyValidatorMgr{}, validatorMgr)
}

func ValidatorMgrFromContext(c context.Context) ValidatorMgr {
	if mgr, ok := c.Value(contextKeyValidatorMgr{}).(ValidatorMgr); ok {
		return mgr
	}
	return ValidatorMgrDefault
}

type ValidatorCreator interface {
	// Names name and aliases of validator
	// we will register validator to validator set by these names
	Names() []string
	// New create new instance
	New(context.Context, *Rule) (Validator, error)
}

type Validator interface {
	// Validate validate value
	Validate(v interface{}) error
	// String stringify validator rule
	String() string
}

func NewValidatorFactory() *ValidatorFactory {
	return &ValidatorFactory{
		validatorSet: map[string]ValidatorCreator{},
	}
}

type ValidatorFactory struct {
	validatorSet map[string]ValidatorCreator
}

func (f *ValidatorFactory) Register(validators ...ValidatorCreator) {
	for i := range validators {
		validator := validators[i]
		for _, name := range validator.Names() {
			f.validatorSet[name] = validator
		}
	}
}

func (f *ValidatorFactory) MustCompile(ctx context.Context, rule []byte, typ typesutil.Type, ruleProcessors ...RuleProcessor) Validator {
	v, err := f.Compile(ctx, rule, typ, ruleProcessors...)
	if err != nil {
		panic(err)
	}
	return v
}

func (f *ValidatorFactory) Compile(ctx context.Context, ruleBytes []byte, typ typesutil.Type, ruleProcessors ...RuleProcessor) (validator Validator, err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	compiled := compiledFromContext(ctx)

	// avoid for tree parse
	if pkgPath := typesutil.Deref(typ).PkgPath(); pkgPath != "" {
		id := pkgPath + "." + typ.Name()

		if v, ok := compiled[id]; ok {
			return v, nil
		}

		compiled[id] = validator
	}

	if len(ruleBytes) == 1 && ruleBytes[0] == '-' {
		ruleBytes = nil
	}

	if len(ruleBytes) == 0 {
		if _, ok := typesutil.EncodingTextMarshalerTypeReplacer(typ); !ok {
			switch typesutil.Deref(typ).Kind() {
			case reflect.Struct:
				ruleBytes = []byte("@struct")
			case reflect.Slice:
				ruleBytes = []byte("@slice")
			case reflect.Map:
				ruleBytes = []byte("@map")
			}
		}
	}

	rule, err := ParseRuleWithType(ruleBytes, typ)
	if err != nil {
		return nil, err
	}

	for i := range ruleProcessors {
		if ruleProcessor := ruleProcessors[i]; ruleProcessor != nil {
			ruleProcessor(rule)
		}
	}

	validatorCreator, ok := f.validatorSet[rule.Name]
	if len(ruleBytes) != 0 && !ok {
		return nil, fmt.Errorf("%s not match any validator", rule.Name)
	}

	return NewValidatorLoader(validatorCreator).New(contextWithCompiled(ContextWithValidatorMgr(ctx, f), compiled), rule)
}

type contextKeyCompiled struct{}

func contextWithCompiled(ctx context.Context, compiled map[string]Validator) context.Context {
	return context.WithValue(ctx, contextKeyCompiled{}, compiled)
}

func compiledFromContext(ctx context.Context) map[string]Validator {
	if c, ok := ctx.Value(contextKeyCompiled{}).(map[string]Validator); ok {
		return c
	}
	return map[string]Validator{}
}
