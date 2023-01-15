## bdk steps i-exec-script

I exec this script in <reference>

### Synopsis

Executes the given script in a shell in the referenced pod and default container.

```
bdk steps i-exec-script [flags]
```

### Examples

```
  
  When I exec this script in pod
    """/bin/bash
    curl localhost:8080/ready
    """

```

### Options

```
  -h, --help   help for i-exec-script
```

### Options inherited from parent commands

```
  -p, --plugins strings   Additional plugin step definitions
```

### SEE ALSO

* [bdk steps](bdk_steps.md)	 - View steps

###### Auto generated by spf13/cobra on 15-Jan-2023