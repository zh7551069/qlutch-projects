@Component
export struct PrimaryButton {
  label: string = ''
  enabled: boolean = true
  onTap: () => void = () => {}
  build() {
    Button(this.label)
      .width('100%')
      .height(50)
      .fontSize(16)
      .fontFamily('NotoSansSC-SemiBold')
      .fontColor('#FFFFFF')
      .backgroundColor(this.enabled ? '#0F172A' : '#64748B')
      .borderRadius(25)
      .enabled(this.enabled)
      .onClick(() => {
        if (this.enabled) {
          this.onTap()
        }
      })
  }
}

@Component
export struct SecondaryButton {
  label: string = ''
  enabled: boolean = true
  onTap: () => void = () => {}
  build() {
    Button(this.label)
      .width('100%')
      .height(50)
      .fontSize(16)
      .fontFamily('NotoSansSC-SemiBold')
      .fontColor('#0F172A')
      .backgroundColor('#FFFFFF')
      .borderRadius(25)
      .border({ width: 1, color: '#E1E8F0' })
      .enabled(this.enabled)
      .onClick(() => {
        if (this.enabled) {
          this.onTap()
        }
      })
  }
}

@Component
export struct BackButton {
  onTap: () => void = () => {}
  build() {
    Button() {
      Text('<')
        .fontSize(18)
        .fontColor('#0F172A')
    }
    .width(40)
    .height(40)
    .backgroundColor('#FFFFFF')
    .borderRadius(20)
    .border({ width: 1, color: '#E2E3E7' })
    .onClick(() => {
      this.onTap()
    })
  }
}

@Component
export struct StepProgress {
  step: number = 1
  build() {
    Stack({ alignContent: Alignment.Start }) {
      Row()
        .width('100%')
        .height(16)
        .borderRadius(100)
        .backgroundColor('#F1F4F9')
      Row()
        .width(`${(Math.min(7, Math.max(0, this.step)) / 7) * 100}%`)
        .height(8)
        .borderRadius(100)
        .linearGradient({
          angle: 90,
          colors: [['#4537CA', 0], ['#6929D8', 0.5], ['#D244EE', 1]]
        })
        .margin({ left: 0, top: 4 })
    }
    .height(16)
    .layoutWeight(1)
  }
}

@Component
export struct ChoiceRow {
  label: string = ''
  selected: boolean = false
  onTap: () => void = () => {}
  build() {
    Row({ space: 8 }) {
      if (this.selected) {
        Text('✓')
          .fontSize(16)
          .fontColor('#0F172A')
      }
      Text(this.label)
        .fontSize(16)
        .fontFamily('NotoSansSC-SemiBold')
        .fontColor('#0F172A')
    }
    .width('100%')
    .height(50)
    .justifyContent(FlexAlign.Center)
    .backgroundColor(this.selected ? '#E8E4FF' : '#FFFFFF')
    .borderRadius(25)
    .shadow({ radius: 8, color: '#140F172A', offsetY: 2 })
    .onClick(() => {
      this.onTap()
    })
  }
}

@Component
export struct PermissionCheck {
  checked: boolean = false
  onTap: () => void = () => {}
  build() {
    Text(this.checked ? '✓' : '')
      .width(24)
      .height(24)
      .fontSize(14)
      .fontColor('#FFFFFF')
      .textAlign(TextAlign.Center)
      .backgroundColor(this.checked ? '#0F172A' : '#FFFFFF')
      .borderRadius(6)
      .border({ width: this.checked ? 0 : 1.5, color: '#E1E8F0' })
      .onClick(() => {
        this.onTap()
      })
  }
}

@Component
export struct OtpCells {
  code: string = ''
  build() {
    Row({ space: 8 }) {
      ForEach([0, 1, 2, 3, 4, 5], (index: number) => {
        Text(index < this.code.length ? this.code.charAt(index) : '')
          .width(48)
          .height(56)
          .fontSize(22)
          .fontFamily('NotoSansSC-SemiBold')
          .fontColor('#0F172A')
          .textAlign(TextAlign.Center)
          .backgroundColor('#FFFFFF')
          .borderRadius(12)
          .border({
            width: this.code.length === index ? 2 : 1,
            color: this.code.length === index ? '#4381E6' : '#E1E8F0'
          })
      }, (index: number) => `${index}`)
    }
  }
}

@Component
export struct ClosedToast {
  visible: boolean = false
  build() {
    if (this.visible) {
      Text('功能尚未开放')
        .fontSize(14)
        .fontFamily('NotoSansSC-SemiBold')
        .fontColor('#FFFFFF')
        .padding({ left: 20, right: 20, top: 12, bottom: 12 })
        .backgroundColor('#0F172A')
        .borderRadius(100)
    }
  }
}

@Component
export struct ClosedChannel {
  label: string = ''
  mark: string = ''
  onTap: () => void = () => {}
  build() {
    Row({ space: 8 }) {
      if (this.mark === '支') {
        Text('支')
          .width(24)
          .height(24)
          .fontSize(14)
          .fontColor('#FFFFFF')
          .textAlign(TextAlign.Center)
          .backgroundColor('#1677FF')
          .borderRadius(5.5)
      } else {
        Circle()
          .width(16)
          .height(16)
          .fill('#22C55E')
      }
      Text(this.label)
        .fontSize(16)
        .fontFamily('NotoSansSC-SemiBold')
        .fontColor('#0F172A')
    }
    .width('100%')
    .height(50)
    .justifyContent(FlexAlign.Center)
    .backgroundColor('#FFFFFF')
    .borderRadius(25)
    .border({ width: 1, color: '#E1E8F0' })
    .onClick(() => {
      this.onTap()
    })
  }
}
