Feature: Store lifecycle against a live OpenFGA server

  Background:
    Given a fresh store with the shared model

  Scenario: Create and read a store back by ID
    When I create a store named "extra"
    Then I can read that store back by ID

  Scenario: Iterating all stores includes created stores
    When I create a store named "extra"
    Then iterating all stores includes the background store and the created store

  Scenario: A deleted store is no longer found
    When I create a store named "temp"
    And I delete that store
    And I get that store
    Then the call fails with a not found error
