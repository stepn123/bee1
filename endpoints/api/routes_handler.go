package management

import (
	"brms/endpoints/logic"
	"brms/endpoints/models"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func Routes(app *fiber.App) {
	app.Post("/insertRuleTemplate", insertRuleTemplate)
	app.Patch("/insertRuletoRuleSet", insertRulestoRuleSet)
	app.Put("/updateRuleSet", updateRuleSet)
	app.Post("/execInput", execInput)
	app.Get("/fetchRules", ListAllRuleSet)
	app.Delete("/deleteRuleSet", deleteRuleSetRoute)
}

func execInput(c *fiber.Ctx) error {
	// Check if method is not POST
	if c.Method() != fiber.MethodPost {
		return fiber.NewError(fiber.StatusMethodNotAllowed, "invalid http method")
	}

	// parsing
	var inputData map[string]interface{}
	if err := c.BodyParser(&inputData); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "The request entity contains invalid or missing data")
	}

	// get rule name from query
	ruleSetName := c.Query("ruleSetName")
	if ruleSetName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "query parameter 'ruleSetName' is required")
	}

	// find rule set
	ruleSet, err := findRuleSetByName(ruleSetName)
	if err != nil {
		if err.Error() == "rule does not exists" {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("rule set '%s' not found", ruleSetName))
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// executing rules
	result, err := logic.Exec(ruleSetName, *ruleSet, inputData)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// print result
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("execute '%s' rule set", ruleSetName),
		"action":  fmt.Sprintf("%v", result),
	})
}

func insertRuleTemplate(c *fiber.Ctx) error {
	// check if method is not post
	if c.Method() != fiber.MethodPost {
		return fiber.NewError(fiber.StatusMethodNotAllowed, "invalid http method")
	}

	// parse to struct
	var ruleSet models.RuleSet
	if err := c.BodyParser(&ruleSet); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "The request entity contains invalid or missing data")
	}

	// validate required fields
	validator := validator.New()
	if err := validator.Struct(ruleSet); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "empty fields")
	}

	// insert ruleSet
	mongoID, err := InsertOneRule(ruleSet)
	if err != nil {
		if err.Error() == "rule set already exists" { // conflict check
			return fiber.NewError(fiber.StatusConflict, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	c.Set("Location", fmt.Sprintf("%s/%s", c.BaseURL(), mongoID)) // set header location to satisfy 201 code

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "new rule set inserted",
	})
}

func insertRulestoRuleSet(c *fiber.Ctx) error {
	// check method
	if c.Method() != fiber.MethodPatch {
		return fiber.NewError(fiber.StatusMethodNotAllowed, "invalid http method")
	}

	// get query params
	ruleSetName := c.Query("ruleSetName")
	if ruleSetName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "query parameter required")
	}

	// parse from request body to struct
	var newRules []models.Rule
	if err := c.BodyParser(&newRules); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "The request entity contains invalid or missing data")
	}

	// insert new rules
	if err := InsertRulestoRuleSet(ruleSetName, newRules); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("%d new rules has been inserted to '%s'", len(newRules), ruleSetName),
	})
}

func updateRuleSet(c *fiber.Ctx) error {
	// Check if method is not PUT
	if c.Method() != fiber.MethodPut {
		return fiber.NewError(fiber.StatusMethodNotAllowed, "invalid http method")
	}

	// Get rule set name from query parameter
	ruleSetName := c.Query("ruleSetName")
	if ruleSetName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "query parameter 'ruleSetName' is required")
	}

	// Parse rule set data from request body
	var updatedRuleSet models.RuleSet
	if err := c.BodyParser(&updatedRuleSet); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, "The request entity contains invalid or missing data")
	}

	// Validate the updated rule set
	validator := validator.New()
	if err := validator.Struct(updatedRuleSet); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid fields in updated rule set")
	}

	// Update the rule set in the database
	if err := UpdateRuleSet(ruleSetName, updatedRuleSet); err != nil {
		if err.Error() == "no data found" {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("no rule set with the give key '%s' exists", ruleSetName))
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("Rule set '%s' has been updated", ruleSetName),
	})
}

func ListAllRuleSet(c *fiber.Ctx) error {
	// check method
	if c.Method() != fiber.MethodGet {
		return fiber.NewError(fiber.StatusMethodNotAllowed, "invalid http method")
	}

	// fetch all rule set
	list, err := FetchAllRules()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Cache-Control", "no-cache")

	// check if no rule set from mongo
	if len(list) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": string("rule set list empty"),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": string("listing all rule sets"),
		"details": list,
	})
}

func deleteRuleSetRoute(c *fiber.Ctx) error {
	// check method
	if c.Method() != fiber.MethodDelete {
		return fiber.NewError(fiber.StatusMethodNotAllowed, "invalid http method")
	}

	// Get rule set name from query parameter
	ruleSetName := c.Query("ruleSetName")
	if ruleSetName == "" {
		return fiber.NewError(fiber.StatusBadRequest, "query parameter 'ruleSetName' is required")
	}

	// Delete the rule set from the database
	if err := deleteRuleSet(ruleSetName); err != nil {
		if err.Error() == "no data exists to be deleted" {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("rule set '%s' has been deleted", ruleSetName),
	})
}
