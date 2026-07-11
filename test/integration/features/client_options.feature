Feature: Client and request options against a live OpenFGA server

  Background:
    Given a fresh store with the shared model

  Scenario: Per-request store and model overrides target a different store
    Given a second store with the shared model granting "user:zoe" "viewer" on "document:secret"
    When I check "user:zoe" "viewer" "document:secret" using the second store and model overrides
    Then the result is allowed

  Scenario: A per-request consistency override succeeds
    Given the tuple "user:anne" "viewer" "document:roadmap" is written
    When I check "user:anne" "viewer" "document:roadmap" with a higher-consistency override
    Then the result is allowed

  Scenario: A custom request header does not break the request
    Given the tuple "user:anne" "viewer" "document:roadmap" is written
    When I check "user:anne" "viewer" "document:roadmap" with a custom header
    Then the result is allowed

  Scenario: A store operation without a store ID fails before any request
    When I read tuples with a client that has no store ID
    Then the call fails because no store ID is set
