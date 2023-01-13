Feature: patching
  @timeout=1m
  Scenario: patching
    Given a resource called cm:
      """
      apiVersion: v1
      kind: ConfigMap
      metadata:
        name: example
        namespace: default
      data:
        foo: bar
      """
    And I create cm
    When I patch cm
      | patch | {"data":{"foo":"nobar"}} |
    Then for at least 10s cm jsonpath '{.data.foo}' should equal nobar
