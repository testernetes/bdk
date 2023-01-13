Feature: basic
  @timeout=1m
  Scenario: within 1 second resource creation
    Given a resource called cm:
    """
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: example
      namespace: default
    """
    And I create cm
    Then within 1s cm jsonpath '{.metadata.uid}' should not be empty
