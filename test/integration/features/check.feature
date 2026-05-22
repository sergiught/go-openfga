Feature: Relationship checks against a live OpenFGA server

  Scenario: A granted relationship is allowed
    Given a fresh store with a document authorization model
    And the tuple "user:anne" "reader" "document:roadmap" is written
    When I check whether "user:anne" has "reader" on "document:roadmap"
    Then the result is allowed

  Scenario: An ungranted relationship is denied
    Given a fresh store with a document authorization model
    When I check whether "user:bob" has "reader" on "document:roadmap"
    Then the result is denied
