name: Feature Request
description: 为 Tortoise 提出新功能建议
title: "[Feature] "
labels: ["enhancement"]
body:
  - type: markdown
    attributes:
      value: |
        ## 功能描述
        请描述您想要的功能。

  - type: textarea
    id: feature-description
    attributes:
      label: 功能描述
      placeholder: 清晰简洁地描述您想要的功能
    validations:
      required: true

  - type: textarea
    id: motivation
    attributes:
      label: 使用场景
      placeholder: 这个功能解决什么问题？适用于哪些场景？
    validations:
      required: true

  - type: textarea
    id: alternatives
    attributes:
      label: 替代方案
      placeholder: 您考虑过哪些替代方案？

  - type: textarea
    id: implementation
    attributes:
      label: 实现建议
      placeholder: |
        - API 设计
        - 数据结构
        - 用户界面

  - type: textarea
    id: additional-context
    attributes:
      label: 其他上下文
      placeholder: |
        - 相关链接
        - 截图
        - 参考资料
