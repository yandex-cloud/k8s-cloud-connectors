# Contributing

(in its current state this file is mostly for myself)

## Code Guidelines

### Naming Convention

1. About variables that represent entities that can exist both in *k8s* and *YC*: if it is in the *YC*, it must be named **resource**.
   Otherwise, it must be named **object**. Short for **object** is **obj**, short for **resource** is **res**.
   If **res** can be confused for **result**, it must be named **resource**.
   
2. If request returns something that we can `wait` for, it is named **operation** or **op**. Otherwise, it is either 
   **response**, or **resp**, or **res**.

### Error Conventions

1. If error is thrown from the external source, such as an SDK, k8s API or some imported library, it must be 
   wrapped into `fmt.Errorf` to explain exactly during which operation it happened. If this happens during reconciliation,
   we must **not** include a name, or a kind of reconciled object, as the logger will take care of that for us.
   Error message must be easily generalized.
   
2. If error is thrown from the internal source, such as utility package or phase, it must be rethrown as it is, 
   without any additional wrapping, as it is already wrapped.