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

  Scenario: Batch check returns per-correlation results
    Given the tuple "user:anne" "viewer" "document:roadmap" is written
    And a batch item "c1" checking "user:anne" has "viewer" on "document:roadmap"
    And a batch item "c2" checking "user:bob" has "viewer" on "document:roadmap"
    When I run the batch check
    Then batch item "c1" is allowed
    And batch item "c2" is denied

  Scenario: Expand returns the userset tree for a relation
    Given the tuple "user:anne" "viewer" "document:roadmap" is written
    When I expand "viewer" on "document:roadmap"
    Then the expansion tree is not empty

  Scenario: List objects returns documents a user can view
    Given the tuple "user:anne" "viewer" "document:roadmap" is written
    And the tuple "user:anne" "viewer" "document:budget" is written
    When I list "document" objects "user:anne" can "viewer"
    Then the objects include "document:roadmap"
    And the objects include "document:budget"

  Scenario: Streamed list objects yields the same documents
    Given the tuple "user:anne" "viewer" "document:roadmap" is written
    When I stream "document" objects "user:anne" can "viewer"
    Then the streamed objects include "document:roadmap"
