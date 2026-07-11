Feature: Assertions round-trip against a live OpenFGA server

  Background:
    Given a fresh store with the shared model

  Scenario: Written assertions are read back
    When I write an assertion that "user:anne" "viewer" "document:roadmap" is expected to be true
    And I read the assertions back
    Then the assertions include "user:anne" "viewer" "document:roadmap" expected true
