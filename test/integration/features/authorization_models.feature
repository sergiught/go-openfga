Feature: Authorization model management against a live OpenFGA server

  Background:
    Given a fresh store with the shared model

  Scenario: Read the model back by ID
    Then I can read the model back by ID

  Scenario: Read the latest model
    Then reading the latest model returns the shared model

  Scenario: List models includes the written model
    Then listing all models includes the written model

  Scenario: Writing an invalid model is a validation error
    When I write an authorization model with an undefined relation reference
    Then the call fails with a validation error
