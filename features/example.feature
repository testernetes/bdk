Feature: basic
  Scenario: within 1 second resource creation
    Given a resource called cm
    """
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: blah
      namespace: default
    data:
      test: test
    """
    When I create cm
      | FieldOwner | matt |
    When I set myvar from cm jsonpath '{.data.test}'
    Then within 1s cm jsonpath '{.data.test}' should equal ${myvar}
