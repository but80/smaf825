# smaf825

SMAFフォーマットの着メロを Arduino + YMF825Board でそれなりに再生するプレイヤーです。

## 準備

1. [Go 1.8 をインストール](https://golang.org/dl/) してください。
2. `go get -u github.com/mersenne-sister/smaf825` にて、当プロジェクトをcloneしてください。
   CLIコマンド `smaf825` が同時にインストールされますので、ためしに引数なしで実行してみてください。
3. 以下のようにして、お手元のMMFファイルが正しくダンプされるかを確認してください。
   
   ```bash
   smaf825 dump music.mmf
   
   smaf825 dump -j music.mmf # JSON形式でもダンプできます
   ```
4. 以下の記事を参考にハードウェアを用意し、まずは記事通りに公式サンプルを鳴らしてみてください。
   - [YMF825BoardをArduinoで鳴らしてみる](https://fabble.cc/yamahafsm/ymf825boardarduino)
5. 無事動作したら、当プロジェクトの [bridge/bridge.ino](bridge/bridge.ino) を参考記事と同様の方法でArduinoに転送してください。

## SMAFの再生

以下のようにして再生できます。
Arduinoを接続しているシリアルポートのデバイス名を指定する必要がありますので、スケッチの転送時に指定したデバイス名をここでも指定してください。

```bash
smaf825 play /dev/tty.usbserial-xxxxxxxx music.mmf
```

## 参考情報

- [YMF825Board GitHubPage](https://yamaha-webmusic.github.io/ymf825board/intro/)

## 注意点

- **すべてのSMAFファイルに対応しているわけではありません**ので、本ツールを用いて特定のSMAFファイルを再生するために
  Arduino や YMF825Board を購入しようとしている方は、ご理解の上でお試しください。
  また、本ツールで読み込み・ダンプできるMMFであっても、以下のような制限により正しく再生されない場合があります。
  - 内蔵波形を用いたFM音色以外（PCMのドラムやユーザ波形等）は再生されません。
  - 1チャンネル内で和音を使用している場合、正しく再生されません。
  - 16和音を超えたチャンネルや、16音色を超えた音色を使用するノートは再生されません。
  - MA-2, MA-5 用のツールで作成したMMFの再生を確認しています。MA-3, MA-7 については未確認です。
- シリアルポート経由でリアルタイムにレジスタへの書き込みを行っています。
  `smaf825` を実行中のマシンで他のプロセスによる負荷がかかっている場合、再生がもたつく場合があります。
