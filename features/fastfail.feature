Feature: Fast failure example
  Scenario: This will fail
    Given a resource called cm:
    """
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: example
      namespace: default
    """
    When I create cm
    Then within 10s cm jsonpath '{.metadata.name}' should equal foo
    And I delete cm
    | propagation policy | Foreground |
  Scenario: this won't run
    Given a resource called cm:
    """
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: example
      namespace: default
    """
    When I create cm
    Then within 10s cm jsonpath '{.metadata.name}' should equal example
    And I delete cm
    | propagation policy | Foreground |
