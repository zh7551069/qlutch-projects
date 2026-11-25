@Component
export struct SplashView {
  onDone: () => void = () => {}
  @State holding: boolean = false
  private left: boolean = false
  private timer: number = -1

  build() {
    Stack() {
      Column() {
        Blank()
        Column({ space: 16 }) {
          Stack() {
            Column()
              .width(260)
              .height(260)
              .borderRadius(22)
              .backgroundColor('#F1F4F9')
            Video({ src: $rawfile('splash.mp4') })
              .width(260)
              .height(260)
              .autoPlay(true)
              .controls(false)
              .objectFit(ImageFit.Contain)
              .borderRadius(22)
              .onFinish(() => {
                this.hold()
              })
          }
          .width(260)
          .height(260)
          Image($rawfile('wordmark.png'))
            .width(143)
            .height(36)
            .objectFit(ImageFit.Contain)
          Text('让记知识，变得轻松有趣')
            .fontSize(15)
            .fontFamily('NotoSansSC-Regular')
            .fontColor('#6C5CE7')
        }
        Blank()
        if (this.holding) {
          Text('正在擦亮你的知识地图')
            .fontSize(15)
            .fontFamily('NotoSansSC-Regular')
            .fontColor('#7C848F')
            .margin({ bottom: 40 })
        } else {
          Blank().height(32)
        }
      }
      .width('100%')
      .height('100%')
      .backgroundColor('#F2F5F8')
    }
    .width('100%')
    .height('100%')
    .onClick(() => {
      this.finish()
    })
  }

  private hold(): void {
    this.holding = true
    this.timer = setTimeout(() => {
      this.finish()
    }, 1000)
  }

  private finish(): void {
    if (this.left) {
      return
    }
    this.left = true
    if (this.timer >= 0) {
      clearTimeout(this.timer)
    }
    this.onDone()
  }
}
