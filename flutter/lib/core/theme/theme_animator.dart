import 'package:flutter/material.dart';

/// 主题动画服务
class ThemeAnimator {
  /// 平滑切换主题
  static void animateThemeChange(
    BuildContext context,
    ThemeMode newMode, {
    Duration duration = const Duration(milliseconds: 300),
    Curve curve = Curves.easeInOut,
  }) {
    Navigator.of(context).pushReplacementNamed('/');
  }

  /// 获取主题切换过渡效果
  static Widget buildThemeTransition({
    required BuildContext context,
    required Widget child,
    required ThemeMode themeMode,
    Duration duration = const Duration(milliseconds: 300),
  }) {
    return AnimatedTheme(
      data: Theme.of(context),
      duration: duration,
      child: child,
    );
  }
}

/// 主题渐变动画
class ThemeGradientTransition extends StatelessWidget {
  final Color startColor;
  final Color endColor;
  final Duration duration;
  final Widget child;

  const ThemeGradientTransition({
    super.key,
    required this.startColor,
    required this.endColor,
    required this.duration,
    required this.child,
  });

  @override
  Widget build(BuildContext context) {
    return TweenAnimationBuilder<Color?>(
      tween: ColorTween(begin: startColor, end: endColor),
      duration: duration,
      builder: (context, color, child) {
        return Container(
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [
                color ?? startColor,
                color?.withAlpha(200) ?? endColor,
              ],
            ),
          ),
          child: child,
        );
      },
      child: child,
    );
  }
}

/// 深色模式过渡动画
class DarkModeTransition extends StatefulWidget {
  final Widget child;
  final bool isDarkMode;
  final Duration duration;

  const DarkModeTransition({
    super.key,
    required this.child,
    required this.isDarkMode,
    this.duration = const Duration(milliseconds: 300),
  });

  @override
  State<DarkModeTransition> createState() => _DarkModeTransitionState();
}

class _DarkModeTransitionState extends State<DarkModeTransition>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _animation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: widget.duration,
    );
    _animation = CurvedAnimation(
      parent: _controller,
      curve: Curves.easeInOut,
    );
    if (widget.isDarkMode) {
      _controller.value = 1.0;
    }
  }

  @override
  void didUpdateWidget(DarkModeTransition oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.isDarkMode != oldWidget.isDarkMode) {
      if (widget.isDarkMode) {
        _controller.forward();
      } else {
        _controller.reverse();
      }
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _animation,
      builder: (context, child) {
        return ColorFiltered(
          colorFilter: ColorFilter.mode(
            Color.lerp(Colors.transparent, Colors.black, _animation.value)!,
            BlendMode.darken,
          ),
          child: widget.child,
        );
      },
    );
  }
}
