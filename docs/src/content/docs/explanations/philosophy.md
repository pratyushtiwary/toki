---
title: Philosophy of Toki
description: This doc goes over the design philosophy Toki follows
---

Toki's desing is guided by three foundational principles that shape every aspect of the library. Understanding these patterns will help you write better configs and make decisions that align with Toki's philosophy.

## Reusability

Toki heavily focus on reusability. The mindset is if you are repeating something more that 2 times across different places it should be made reusable. Toki allows reusability by `inherit` strategy. As a user of Toki your focus should be how much of your steps you can put in project config rather than pipeline config. Each step should start from pipeline config but the ultimate goal for it should be to reach project config.

Reusability also allows you to create and share util project configs, some pain points or configuration that takes lot of time should be abstracted away using project configs.

## Separation of concern

Toki achieves separation of concern(SOC) by abstracting away the token logic from your program. In an ideal world your program shouldn't even care about token's freshness, it should focus on other goals such as processing data or querying a database. SOC also allows your programs to be flexible, if today your program queries dynamodb and in future it wants to query redshit as well then you(the programmer) shouldn't think about tokens and auth much and rather you should focus more on the core aspect of the task.

SOC can be seen in Toki's codebase as well, each module has a separate concern and Toki's main entrypoint orchestrates these modules to work correctly.

## Batteries Included, But Swappable

Toki covers most of the use-cases of token and re-auth. This is done via 3 foundational strategies:
- `curl_cookie`,
- `custom`,
- `inherit`

If `curl_cookie` or `inherit` doesn't fit your use-case you are always welcome to use `custom` strategy, as almost all the auth refresh commands can be configured via `custom` strategy.

For developers, Toki's separation of concern philosophy and swappable philosophy makes it easier to add a new strategy and pushing to Toki.