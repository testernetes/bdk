Feature: Logging
  Scenario: BDK Should assert on container logs
    Given a resource called pod:
    """yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: app
      namespace: default
    spec:
      restartPolicy: Never
      containers:
      - command: ["busybox", "httpd", "-f", "-p", "8000"]
        image: busybox:latest
        name: server
    """
    When I create pod
    And within 1m pod jsonpath '{.status.phase}' should equal Running
    And I proxy get http://pod:8000/fake
