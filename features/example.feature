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
    When I create cm
    Then within 10s cm jsonpath '{.metadata.name}' should equal example
    And I delete cm
    | propagation policy | Foreground |

