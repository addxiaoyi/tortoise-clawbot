name: Bug Report
description: 报告一个 bug 帮助我们改进
title: "[Bug] "
labels: ["bug"]
body:
  - type: markdown
    attributes:
      value: |
        ## 描述
        请简要描述这个 bug。

  - type: textarea
    id: bug-description
    attributes:
      label: Bug 描述
      placeholder: 详细描述 bug 的症状
    validations:
      required: true

  - type: textarea
    id: steps
    attributes:
      label: 复现步骤
      placeholder: |
        1. Go to '...'
        2. Click on '...'
        3. Scroll down to '...'
        4. See error
    validations:
      required: true

  - type: textarea
    id: expected
    attributes:
      label: 预期行为
      placeholder: 描述您期望发生的行为

  - type: textarea
    id: logs
    attributes:
      label: 日志/错误信息
      placeholder: |
        ```
        Paste any relevant log output here
        ```

  - type: dropdown
    id: os
    attributes:
      label: 操作系统
      options:
        - Linux
        - macOS
        - Windows
        - Android
        - iOS
        - Other
    validations:
      required: true

  - type: input
    id: version
    attributes:
      label: Tortoise 版本
      placeholder: 0.1.0
