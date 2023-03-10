## bdk steps it-should-log-duration

<assertion> <duration> <reference> logs (should|should not) say <text>

### Synopsis

Asserts that the referenced resource will log something within the specified duration

```
bdk steps it-should-log-duration [flags]
```

### Examples

```
  
  Given a resource called testernetes:
    """
    apiVersion: v1
    kind: Pod
    metadata:
      name: testernetes
      namespace: default
    spec:
      restartPolicy: Never
      containers:
      - command:
        - /bdk
        - --help
        image: ghcr.io/testernetes/bdk:d408c829f019f2052badcb93282ee6bd3cdaf8d0
        name: bdk
    """
  When I create testernetes
  Then within 1m testernetes logs should say Behaviour Driven Kubernetes
    | container | bdk   |
    | follow    | false |

```

### Options

```
  -h, --help   help for it-should-log-duration
```

### Options inherited from parent commands

```
  -p, --plugins strings   Additional plugin step definitions
```

### SEE ALSO

* [bdk steps](bdk_steps.md)	 - View steps

###### Auto generated by spf13/cobra on 15-Jan-2023
