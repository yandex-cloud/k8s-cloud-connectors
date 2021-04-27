# Notice to external contributors


## General info

Hello! In order for us (YANDEX LLC) to accept patches and other contributions from you, you will have to adopt our Yandex Contributor License Agreement (the "**CLA**"). The current version of the CLA can be found here:
1) https://yandex.ru/legal/cla/?lang=en (in English) and
2) https://yandex.ru/legal/cla/?lang=ru (in Russian).

By adopting the CLA, you state the following:

* You obviously wish and are willingly licensing your contributions to us for our open source projects under the terms of the CLA,
* You have read the terms and conditions of the CLA and agree with them in full,
* You are legally able to provide and license your contributions as stated,
* We may use your contributions for our open source projects and for any other our project too,
* We rely on your assurances concerning the rights of third parties in relation to your contributions.

If you agree with these principles, please read and adopt our CLA. By providing us your contributions, you hereby declare that you have already read and adopt our CLA, and we may freely merge your contributions with our corresponding open source project and use it in further in accordance with terms and conditions of the CLA.

## Provide contributions

If you have already adopted terms and conditions of the CLA, you are able to provide your contributions. When you submit your pull request, please add the following information into it:

```
I hereby agree to the terms of the CLA available at: [link].
```

Replace the bracketed text as follows:
* [link] is the link to the current version of the CLA: https://yandex.ru/legal/cla/?lang=en (in English) or https://yandex.ru/legal/cla/?lang=ru (in Russian).

It is enough to provide us such notification once.

## Other questions

If you have any questions, please mail us at opensource@yandex-team.ru.

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