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

  Scenario: Viewer is inherited from editor
    Given the tuple "user:anne" "editor" "document:roadmap" is written
    When I check whether "user:anne" has "viewer" on "document:roadmap"
    Then the result is allowed

  Scenario: Viewer is inherited from owner
    Given the tuple "user:anne" "owner" "document:roadmap" is written
    When I check whether "user:anne" has "viewer" on "document:roadmap"
    Then the result is allowed

  Scenario: Viewer is granted through group membership
    Given the tuple "user:anne" "member" "group:eng" is written
    And the tuple "group:eng#member" "editor" "document:roadmap" is written
    When I check whether "user:anne" has "viewer" on "document:roadmap"
    Then the result is allowed

  Scenario: A contextual tuple grants access for one check
    When I check whether "user:anne" has "viewer" on "document:roadmap" with a contextual tuple "user:anne" "viewer" "document:roadmap"
    Then the result is allowed

  Scenario: Checking an unknown relation is a validation error
    When I check whether "user:anne" has "bogus" on "document:roadmap"
    Then the call fails with a validation error
