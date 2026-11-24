#!/usr/bin/env bash
# 一键启动 Android 模拟器并看到 Qlutch 的启动页 / 欢迎页。
#
# 为什么要有这个脚本：双击 Finder 里的 .command 时不会加载 ~/.zshrc，
# 所以 ANDROID_HOME / JAVA_HOME 都得自己找，不能指望调用者的环境。
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
ANDROID_DIR="$REPO_ROOT/apps/android"
APP_ID="dev.qlutch.app"
SPLASH_ACT="$APP_ID/.splash.SplashActivity"
DEFAULT_AVD="Medium_Phone_API_37.0"
EMU_LOG="${TMPDIR:-/tmp}/qlutch-emulator.log"
BOOT_TIMEOUT=240

AVD="$DEFAULT_AVD"
COLD=0
TARGET="splash"
SHOT=""
RUN_TESTS=0
DO_BUILD=1

info()  { printf '\033[36m==>\033[0m %s\n' "$*"; }
warn()  { printf '\033[33m!!\033[0m %s\n' "$*" >&2; }
die()   { printf '\033[31mERR\033[0m %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<USAGE
用法: tools/android/demo.sh [选项]

  --avd NAME    指定 AVD（默认 $DEFAULT_AVD），--list 可看全部
  --cold        冷启动模拟器（已在跑则重启，且不加载快照）
  --welcome     直接进欢迎页（跳过启动动画）
  --splash      只走启动页（默认）
  --shot PATH   跑完截图存到 PATH
  --tests       先跑单元测试再装
  --no-build    跳过构建，只启动已安装的版本
  --list        列出可用 AVD 后退出
  -h, --help    显示本帮助

例子:
  tools/android/demo.sh                     # 起模拟器 + 编译安装 + 看启动页
  tools/android/demo.sh --welcome --shot /tmp/w.png
  tools/android/demo.sh --cold --tests
USAGE
}

while [ $# -gt 0 ]; do
  case "$1" in
    --avd)      AVD="$2"; shift 2 ;;
    --cold)     COLD=1; shift ;;
    --welcome)  TARGET="welcome"; shift ;;
    --splash)   TARGET="splash"; shift ;;
    --shot)     SHOT="$2"; shift 2 ;;
    --tests)    RUN_TESTS=1; shift ;;
    --no-build) DO_BUILD=0; shift ;;
    --list)     LIST_ONLY=1; shift ;;
    -h|--help)  usage; exit 0 ;;
    *)          die "未知参数: $1（--help 看用法）" ;;
  esac
done

# ---------- 定位 SDK / JDK ----------
SDK="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-$HOME/Library/Android/sdk}}"
[ -d "$SDK" ] || die "找不到 Android SDK：$SDK（可设 ANDROID_HOME）"
EMULATOR="$SDK/emulator/emulator"
ADB="$SDK/platform-tools/adb"
[ -x "$ADB" ] || die "找不到 adb：$ADB"
export ANDROID_HOME="$SDK"
export ANDROID_SDK_ROOT="$SDK"
export PATH="$SDK/platform-tools:$SDK/emulator:$PATH"

if [ -n "${JAVA_HOME:-}" ] && [ -x "$JAVA_HOME/bin/java" ]; then
  :
elif [ -x "/Applications/Android Studio.app/Contents/jbr/Contents/Home/bin/java" ]; then
  export JAVA_HOME="/Applications/Android Studio.app/Contents/jbr/Contents/Home"
else
  JH="$(/usr/libexec/java_home -v 17 2>/dev/null || /usr/libexec/java_home 2>/dev/null || true)"
  [ -n "$JH" ] && export JAVA_HOME="$JH"
fi
[ -n "${JAVA_HOME:-}" ] && [ -x "$JAVA_HOME/bin/java" ] \
  || die "找不到 JDK（装 Android Studio，或设置 JAVA_HOME）"

if [ "${LIST_ONLY:-0}" = "1" ]; then
  info "可用 AVD："
  "$EMULATOR" -list-avds || true
  exit 0
fi

# ---------- 模拟器 ----------
running_device() {
  "$ADB" devices | awk '$2=="device" && $1 ~ /^emulator-/ {print $1; exit}'
}

booted() {
  [ "$("$ADB" -s "$1" shell getprop sys.boot_completed 2>/dev/null | tr -d '\r')" = "1" ]
}

start_emulator() {
  [ -x "$EMULATOR" ] || die "找不到 emulator：$EMULATOR"
  local known
  known="$("$EMULATOR" -list-avds || true)"
  if ! printf '%s\n' "$known" | grep -qx "$AVD"; then
    warn "AVD 列表："
    printf '%s\n' "$known" | sed 's/^/    /' >&2
    die "没有名为 $AVD 的 AVD（--avd 换一个）"
  fi

  local args=(-avd "$AVD" -netdelay none -netspeed full -no-boot-anim)
  if [ "$COLD" = "1" ]; then
    args+=(-no-snapshot-load)
    info "冷启动（不加载快照）"
  fi

  info "启动模拟器 $AVD …（日志 $EMU_LOG）"
  nohup "$EMULATOR" "${args[@]}" >"$EMU_LOG" 2>&1 &
  disown 2>/dev/null || true
}

DEVICE="$(running_device || true)"
if [ -n "$DEVICE" ] && [ "$COLD" = "1" ]; then
  info "按 --cold 要求重启 $DEVICE"
  "$ADB" -s "$DEVICE" emu kill >/dev/null 2>&1 || true
  for _ in $(seq 1 40); do
    [ -z "$(running_device || true)" ] && break
    sleep 0.5
  done
  DEVICE=""
fi

if [ -z "$DEVICE" ]; then
  start_emulator
  info "等待设备上线 …"
  "$ADB" wait-for-device
  DEVICE="$("$ADB" devices | awk '$2=="device" && $1 ~ /^emulator-/ {print $1; exit}')"
  [ -n "$DEVICE" ] || die "模拟器没起来，看 $EMU_LOG"
else
  info "复用已在运行的 $DEVICE"
fi

if ! booted "$DEVICE"; then
  info "等待系统启动完成（最多 ${BOOT_TIMEOUT}s）…"
  waited=0
  until booted "$DEVICE"; do
    sleep 2; waited=$((waited+2))
    [ "$waited" -ge "$BOOT_TIMEOUT" ] && die "启动超时，看 $EMU_LOG"
    [ $((waited % 20)) -eq 0 ] && info "  ${waited}s …"
  done
fi
info "设备就绪：$DEVICE"

# 唤醒并解锁，否则截图是熄屏
"$ADB" -s "$DEVICE" shell input keyevent 224 >/dev/null 2>&1 || true
"$ADB" -s "$DEVICE" shell wm dismiss-keyguard >/dev/null 2>&1 || true

# ---------- 构建安装 ----------
cd "$ANDROID_DIR"
CHMOD_HINT="chmod +x $ANDROID_DIR/gradlew"
[ -x ./gradlew ] || { warn "gradlew 没有执行权限，正在修复"; chmod +x ./gradlew 2>/dev/null || die "请手动执行 $CHMOD_HINT"; }

if [ "$RUN_TESTS" = "1" ]; then
  info "运行单元测试 …"
  ./gradlew --console=plain :app:testDebugUnitTest
fi

if [ "$DO_BUILD" = "1" ]; then
  info "编译并安装 Debug 包 …"
  ./gradlew --console=plain :app:installDebug
else
  info "--no-build：跳过编译安装"
fi

# ---------- 拉起 App ----------
pkg_installed() {
  "$ADB" -s "$DEVICE" shell pm list packages 2>/dev/null | tr -d '\r' | grep -qx "package:$APP_ID"
}
pkg_installed || die "设备上没有 $APP_ID，去掉 --no-build 重新来一次"

# ---------- 等前台稳定 ----------
top_activity() {
  "$ADB" -s "$DEVICE" shell dumpsys activity activities 2>/dev/null \
    | grep -m1 'topResumedActivity' \
    | grep -oE '[A-Za-z0-9_.]+/[A-Za-z0-9_.$]+' | head -1 | sed 's|.*[/.]||' || true
}

# WelcomeActivity 没有 exported，adb 起不来；启动页本身支持点击跳过，
# 所以 --welcome 的做法是照常起 SplashActivity，再点一下屏幕跳过动画。
WANT="SplashActivity"
[ "$TARGET" = "welcome" ] && WANT="WelcomeActivity"

info "停止旧进程"
"$ADB" -s "$DEVICE" shell am force-stop "$APP_ID" >/dev/null 2>&1 || true

info "启动应用 …"
START_OUT="$("$ADB" -s "$DEVICE" shell am start -n "$SPLASH_ACT" 2>&1 || true)"
case "$START_OUT" in
  *Error*|*Exception*) die "启动失败：$START_OUT" ;;
esac

if [ "$TARGET" = "welcome" ]; then
  # 等启动页真正到前台，再点一下跳过动画
  for _ in $(seq 1 40); do
    [ "$(top_activity || true)" = "SplashActivity" ] && break
    sleep 0.25
  done
  sleep 0.6
  SIZE="$("$ADB" -s "$DEVICE" shell wm size 2>/dev/null | tr -d '\r' | sed -n 's/.*: //p')"
  W="${SIZE%x*}"; H="${SIZE#*x}"
  [ -n "$W" ] && [ -n "$H" ] || { W=1080; H=2400; }
  info "点击屏幕跳过启动动画 → 欢迎页"
  "$ADB" -s "$DEVICE" shell input tap $((W/2)) $((H/2)) >/dev/null 2>&1 || true
fi

# ---------- 等前台稳定 ----------
waited=0
while [ "$waited" -lt 40 ]; do
  CUR="$(top_activity || true)"
  [ "$CUR" = "$WANT" ] && break
  sleep 0.5; waited=$((waited+1))
done
CUR="$(top_activity || true)"
if [ "$CUR" = "$WANT" ]; then
  info "前台已切换到 $WANT"
else
  warn "前台当前是 ${CUR:-未知}，期望 $WANT —— 可能是还在播启动动画"
fi

# 让动画跑一帧，截图不会是空白
sleep 1

if [ -n "$SHOT" ]; then
  mkdir -p "$(dirname "$SHOT")"
  "$ADB" -s "$DEVICE" exec-out screencap -p > "$SHOT"
  info "截图已存：$SHOT"
fi

info "完成。窗口应该已经能看到效果了；关掉窗口用 'adb emu kill' 或直接关模拟器窗口。"
