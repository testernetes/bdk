Feature: Logging
  Scenario: BDK Should assert on container logs
    Given a resource called pod
    """yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: test
      namespace: default
    spec:
      restartPolicy: Never
      containers:
      - command:
        - sleep
        - "300"
        image: ghcr.io/testernetes/bdk:d408c829f019f2052badcb93282ee6bd3cdaf8d0
        name: bdk
    """
    When I create pod
    Then I evict pod
