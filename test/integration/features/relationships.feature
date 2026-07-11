Feature: Relationship queries against a live OpenFGA server

  Background:
    Given a fresh store with the shared model

  Scenario: A granted relationship is allowed
    Given the tuple "user:anne" "viewer" "document:roadmap" is written
    When I check whether "user:anne" has "viewer" on "document:roadmap"
    Then the result is allowed

  Scenario: An ungranted relationship is denied
    When I check whether "user:bob" has "viewer" on "document:roadmap"
    Then the result is denied
