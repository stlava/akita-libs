Allows users to specify path generalization.

Example:

```
Input: /v1/{my_path_param}

Original endpoint   Post-processed endpoint
/v1/foo         /v1/{my_path_param}
/v1/x/y         /v1/{my_path_param}/y
/v1/{arg2}/z    /v1/{my_path_param}/z
```
