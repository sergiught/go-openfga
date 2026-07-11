Feature: Tuple writes and reads against a live OpenFGA server

  Background:
    Given a fresh store with the shared model

  Scenario: Written tuples are read back
    Given the tuple "user:anne" "viewer" "document:roadmap" is written
    When I read tuples for object "document:roadmap"
    Then the tuples include "user:anne" "viewer" "document:roadmap"

  Scenario: Deleted tuples are gone
    Given the tuple "user:anne" "viewer" "document:roadmap" is written
    When I delete the tuple "user:anne" "viewer" "document:roadmap"
    And I read tuples for object "document:roadmap"
    Then the tuples do not include "user:anne" "viewer" "document:roadmap"

  Scenario: The changelog records writes
    Given the tuple "user:anne" "viewer" "document:roadmap" is written
    When I read all tuple changes
    Then the changes include a "TUPLE_OPERATION_WRITE" of "user:anne" "viewer" "document:roadmap"

  Scenario: Writing a tuple with an unknown relation is a validation error
    When I write the tuple "user:anne" "bogus" "document:roadmap"
    Then the call fails with a validation error
