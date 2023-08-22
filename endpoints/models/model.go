package models

type Rule struct {
	Id         int                    `bson:"id" json:"id"`
	Conditions map[string]interface{} `bson:"conditions" json:"conditions"`
	Action     interface{}            `bson:"action" json:"action"`
}

type ConditionDictionary struct {
	Label     string `bson:"label" json:"label"`
	Attribute string `bson:"attribute" json:"attribute"`
	Operator  string `bson:"operator" json:"operator"`
}

type Body struct {
	Name string `bson:"name" json:"name"`
	Type string `bson:"type" json:"type"`
}

type Action struct {
	Label     string `bson:"label" json:"label"`
	Attribute string `bson:"attribute" json:"attribute"`
	Type      string `bson:"type" json:"type"`
}

type Description struct {
	Condition string `bson:"condition" json:"condition"`
	Action    string `bson:"action" json:"action"`
}

type RuleSet struct {
	Name        string                `bson:"name" json:"name" validate:"required"`
	Endpoint    string                `bson:"endpoint" json:"endpoint" validate:"required"`
	Bodies      []Body                `bson:"bodies" json:"bodies" validate:"required"`
	Conditions  []ConditionDictionary `bson:"conditions" json:"conditions" validate:"required"`
	Action      `bson:"action" json:"action" validate:"required"`
	Rules       []Rule `bson:"rules" json:"rules"`
	Description `bson:"description" json:"description"`
}
