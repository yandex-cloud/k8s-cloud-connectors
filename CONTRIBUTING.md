# Notice to external contributors

## General info

Hello! In order for us (YANDEX LLC) to accept patches and other contributions from you, you will have to adopt our
Yandex Contributor License Agreement (the "**CLA**"). The current version of the CLA can be found here:

1) https://yandex.ru/legal/cla/?lang=en (in English) and
2) https://yandex.ru/legal/cla/?lang=ru (in Russian).

By adopting the CLA, you state the following:

* You obviously wish and are willingly licensing your contributions to us for our open source projects under the terms
  of the CLA,
* You have read the terms and conditions of the CLA and agree with them in full,
* You are legally able to provide and license your contributions as stated,
* We may use your contributions for our open source projects and for any other our project too,
* We rely on your assurances concerning the rights of third parties in relation to your contributions.

If you agree with these principles, please read and adopt our CLA. By providing us your contributions, you hereby
declare that you have already read and adopt our CLA, and we may freely merge your contributions with our corresponding
open source project and use it in further in accordance with terms and conditions of the CLA.

## Provide contributions

If you have already adopted terms and conditions of the CLA, you are able to provide your contributions. When you submit
your pull request, please add the following information into it:

```
I hereby agree to the terms of the CLA available at: [link].
```

Replace the bracketed text as follows:

* [link] is the link to the current version of the CLA: https://yandex.ru/legal/cla/?lang=en (in English)
  or https://yandex.ru/legal/cla/?lang=ru (in Russian).

It is enough to provide us such notification once.

## Other questions

If you have any questions, please mail us at opensource@yandex-team.ru.

## Contribution guidelines

### How to create new connector
Let's follow a step-by-step guide for creation of **YetAnotherResource** connector.

1. You can use [scaffolder](scaffolder) to create baseline connector for you in a `./connector` directory, or 
   you can just copy-paste one of the existing ones and refactor it a bit. 
2. Add the resource api to the manager scheme by adding `utilruntime.Must(yar.AddToScheme(scheme))` line
   (where `yar` is name of the `api/v1` import of your resource folder) into 
   [main's](cmd/yc-connector-manager/main.go) `init` function.
3. In the `execute` function of the same file setup that connector
   (you can look at how it is done for existing ones and do it in the same manner), passing down everything needed,
   such as *logger*, *context* and *clusterId*.
4. You're ready to use your connector, now it is time to fill it with some code!

### How to create new webhook

1. In order to create new webhook, you shall create a struct that implements either `webhook.Validator` or `webhook.Mutator`,
   depending on the type of webhook needed.
2. Add a marker comment specifying your webhook configuration. You can use [this webhook](connector/ycr/webhook/validating.go) as
   an example, or you can look for detailed description of each parameter [here](https://book.kubebuilder.io/reference/markers/webhook.html).
3. Register your webhook in the [manager](cmd/yc-connector-manager/main.go) `execute` function, following existing ones
   as an example.
4. You're good to go!

#### Things to remember
* You must pass an empty object of the type that webhook supervises into the helper functions 
  `webhook.RegisterValidatingHandler` and `webhook.RegisterMutatingHandler`. This is used to unmarshall **k8s**
  requests into (Go 2.0 and Generics are near, but not here yet).
* In every method of `webhook.Validator` or `webhook.Mutator` you get a raw object, so you should cast it to desired
  type by hands. It is better to use *optimistic* casts without typechecking, because there is a layer in webhook system that
  recovers from panics, just in case something goes wrong.

## Code Guidelines

### Naming Convention

* About variables that represent entities that can exist both in *k8s* and *YC*: if it is in the *YC*, it must be
   named **resource**. Otherwise, it must be named **object**. Short for **object** is **obj**, short for **resource**
   is **res**. If **res** can be confused for **result**, it must be named **resource**.

* If request returns something that we can `wait` for, it is named **operation** or **op**. Otherwise, it is either
   **response**, or **resp**, or **res**.

### Error Conventions

* If error is thrown from the external source, such as an SDK, k8s API or some imported library, it must be wrapped
   into `fmt.Errorf` to explain exactly during which operation it happened. If this happens during reconciliation, we
   must **not** include a name, or a kind of reconciled object, as the logger will take care of that for us. Error
   message must be easily generalized.

* Errors thrown during reconciliation process must not be logged, because they would be logged when reconciler throws it,
   though it must be wrapped with `fmt.Errorf` to understand context of error.
  
* If your code receives an `error` from a subordinate function call and is not re-throwing it, it **must** be logged
   on at least `ERROR` level.
  
### Logging conventions

* Remember to pass logger to any underlying function specifying that function via `.WithName(string)`. Thus, 
  it would be possible to write very simple logging messages, such as "started" or "finished", as well as writing
  common code that accepts logger as an argument without specifying, where it is called 
  (for example, `phase.RegisterFinalizer` can be called from both `YandexContainerRegistryReconciler` 
  and `YandexObjectStorage`, and inner logging would remain the same).
* While most loggers have debug-level logging via some method like `log.Debug(message)`, unfortunately, 
  `logr.Logger` does not have this method. To imitate it, you shall use `log.V(1).Info(message)`.