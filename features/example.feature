Feature: basic
  Scenario: within 1 second resource creation
    Given a resource called cm
    """
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: ${TESTING}
      namespace: default
    """
    When I create cm
    Then within 10s cm jsonpath '{.metadata.name}' should equal ${TESTING}1
