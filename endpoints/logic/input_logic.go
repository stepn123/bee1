package logic

import (
	"brms/endpoints/models"
	"fmt"

	"github.com/Knetic/govaluate"
)

func ValidateRule(operator string, initValue, inputValue interface{}) (bool, error) {
	expressionString := fmt.Sprintf("%v %s %v", inputValue, operator, initValue)
	expression, err := govaluate.NewEvaluableExpression(expressionString)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	parameters := make(map[string]interface{})
	result, err := expression.Evaluate(parameters)
	if err != nil {
		return false, err
	}

	return result.(bool), nil
}

func CheckBoolean(slice []bool) bool {
	for _, value := range slice {
		if !value {
			return false
		}
	}
	return true
}

func Exec(ruleSet string, rulesSelected models.RuleSet, inputUSer map[string]interface{}) (interface{}, error) {
	var boolComplete []bool
	for _, rule := range rulesSelected.Rules {
		boolComplete = boolComplete[:0] // emptying the slice
		for _, condition := range rulesSelected.Conditions {
			// cek empty map
			if len(rule.Conditions) == 0 {
				return rule.Action, nil
			}
			// cek key ada apa kgk
			if _, exists := rule.Conditions[condition.Label]; !exists {
				continue
			}
			// cek key di inout user ada apa kgk
			if inputUSer[condition.Attribute] == nil {
				boolComplete = append(boolComplete, false)
				continue
			}
			// validasi rule
			result, err := ValidateRule(condition.Operator, rule.Conditions[condition.Label], inputUSer[condition.Attribute])
			if err != nil {
				return nil, err
			}
			boolComplete = append(boolComplete, result)
		}
		if CheckBoolean(boolComplete) {
			return rule.Action, nil
		}
	}

	return nil, nil
}
