import { PrimaryButton, SecondaryButton } from '../ui/Controls'

@Component
export struct WelcomeView {
  onStart: () => void = () => {}
  onExisting: () => void = () => {}
  @State frame: number = 1
  private timer: number = -1

  aboutToAppear(): void {
    this.timer = setInterval(() => {
      this.frame = this.frame >= 50 ? 1 : this.frame + 1
    }, 40)
  }

  aboutToDisappear(): void {
    if (this.timer >= 0) {
      clearInterval(this.timer)
    }
  }

  build() {
    Column() {
      Column({ space: 20 }) {
        Stack() {
          Column()
            .width(358)
            .height(240)
            .borderRadius(22)
            .backgroundColor('#F1F4F9')
          Image($rawfile(`cheer/welcome_cheer_${this.padded()}.png`))
            .width(358)
            .height(240)
            .objectFit(ImageFit.Contain)
            .offset({ x: -7, y: 20 })
        }
        .width(358)
        .height(240)
        .borderRadius(22)
        .clip(true)
        Image($rawfile('wordmark.png'))
          .width(174)
          .height(44)
          .objectFit(ImageFit.Contain)
        Text('沉迷于学习，对学习上瘾')
          .width(358)
          .fontSize(16)
          .fontFamily('NotoSansSC-SemiBold')
          .fontColor('#0F172A')
          .textAlign(TextAlign.Center)
      }
      .layoutWeight(1)
      .justifyContent(FlexAlign.Center)
      .padding({ left: 16, right: 16, top: 24 })
      Column({ space: 16 }) {
        PrimaryButton({ label: '开始', onTap: () => { this.onStart() } })
        SecondaryButton({ label: '已有账号', onTap: () => { this.onExisting() } })
        Text('教材不用于训练模型')
          .fontSize(13)
          .fontFamily('NotoSansSC-Regular')
          .fontColor('#7C848F')
        Text('继续即表示同意《用户协议》和《隐私政策》')
          .width(358)
          .fontSize(13)
          .fontFamily('NotoSansSC-Regular')
          .fontColor('#7C848F')
          .textAlign(TextAlign.Center)
      }
      .padding({ left: 16, right: 16, bottom: 28 })
    }
    .width('100%')
    .height('100%')
    .backgroundColor('#F2F5F7')
  }

  private padded(): string {
    if (this.frame < 10) {
      return `00${this.frame}`
    }
    if (this.frame < 100) {
      return `0${this.frame}`
    }
    return `${this.frame}`
  }
}
