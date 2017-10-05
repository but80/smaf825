# smaf825

SMAFフォーマットの着メロを Arduino + YMF825Board で再生するプレイヤーです。

## 準備

1. 以下の記事を参考にハードウェアを用意し、まずは公式サンプルを鳴らしてみてください。
   - [YMF825BoardをArduinoで鳴らしてみる](https://fabble.cc/yamahafsm/ymf825boardarduino)
2. 無事動作したら、当プロジェクトの [bridge/bridge.ino](bridge/bridge.ino) を参考記事と同様の方法でArduinoに転送してください。
3. [Go 1.8 をインストール](https://golang.org/dl/) してください。
4. `go get -u github.com/mersenne-sister/smaf825` にて、当プロジェクトのCLIをインストールしてください。

## SMAFの再生

以下のようにして再生できます。
Arduinoを接続しているシリアルポートのデバイス名を指定する必要がありますので、
スケッチの転送時に指定したデバイス名をここでも指定してください。

```bash
smaf825 play mmf /dev/tty.usbserial-xxxxxxxx music.mmf
```

## 参考情報

- [YMF825Board GitHubPage](https://yamaha-webmusic.github.io/ymf825board/intro/)

## 注意点

- MA-2, MA-5 用のツールで作成したMMFの再生を確認しています。MA-3, MA-7 については未確認です。
- 内蔵波形を用いたFM音色以外（PCMのドラムやユーザ波形等）は再生されません。
- 1チャンネル内で和音を使用している場合、正しく再生されません。
- 16和音を超えたチャンネルや、16音色を超えた音色を使用するノートは再生されません。
- SPFフォーマットには未対応です。
- シリアルポート経由でリアルタイムにレジスタへの書き込みを行っています。
  `smaf825` を実行中のマシンで他のプロセスによる負荷がかかっている場合、再生がもたつく場合があります。
