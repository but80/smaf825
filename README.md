# smaf825

SMAFフォーマットの着メロを Arduino + YMF825Board でそれなりに再生するプレイヤーです。

- macOS Sierra、Windows 10で動作確認しています。
- `v1.2.0` にてバッファリング再生を実装しました。Windowsで特に大きかったもたつきが解消されています。
- 後述の **注意点** をお読みください。

## 準備 (1/2)

以下のいずれかの手順で、CLIコマンド `smaf825` をインストールしてください。

### macOS+Homebrew でのインストール

```bash
brew tap but80/tap
brew install smaf825
```

### 自前でインストール

1. [Releases](https://github.com/but80/smaf825/releases) からアーカイブをダウンロードしてください。
2. アーカイブに含まれるバイナリを適当な場所に置き、パスを通してください。

### Go 1.8 でのインストール（開発者向け）

1. [Go 1.8 をインストール](https://golang.org/dl/) してください。
2. `go get -u github.com/but80/smaf825` にて、当プロジェクトをcloneしてください。

注意：こちらの手順でインストールされた `smaf825` は、バージョン番号を表示しません。

## 準備 (2/2)

1. 以下のようにして、お手元のMMFファイルが正しくダンプされるかを確認してください。
   
   ```bash
   smaf825 dump music.mmf
   
   smaf825 dump -j music.mmf # JSON形式でもダンプできます
   ```
2. 以下の記事を参考にハードウェアを用意し、まずは記事通りに公式サンプルを鳴らしてみてください。
   - [YMF825BoardをArduinoで鳴らしてみる](https://fabble.cc/yamahafsm/ymf825boardarduino)
3. 無事動作したら、当プロジェクトの [bridge/bridge.ino](bridge/bridge.ino) （バージョンが一致するもの）を参考記事と同様の方法でArduinoに転送してください。

## SMAFの再生

以下のようにして再生できます。
Arduinoを接続しているシリアルポートのデバイス名を指定する必要がありますので、スケッチの転送時に指定したデバイス名をここでも指定してください。

```bash
# macOS
smaf825 play /dev/tty.usbserial-xxxxxxxx music.mmf

# Windows
smaf825 play comX music.mmf
```

以下のようにオプションを与えることで、音量を調整できます。

```bash
# -g: Gain (0..3, default=1)
# -v: Analog Volume (0..63, default=48)
smaf825 play -g 3 -v 63 /dev/tty.usbserial-xxxxxxxx music.mmf
```

## YMF825用トーンデータの抽出

`smaf825 dump -v music.mmf` で、MMFやSPFからトーンデータのみを抽出できます。

SMF825用に変換されたデータ列は、JSON形式の出力中の `.voices[].ymf825_data` に含まれます。
[jq](https://stedolan.github.io/jq/) を用いると以下のように取り出せます。

```bash
smaf825 dump -Q -v -j music.mmf | jq -crM '.voices[].ymf825_data'
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
  - MA-7用に作成されたSMAFファイルには未対応です。
- 再生中に `Ctrl+C` で停止後、再度再生しようとすると応答がなくなる・音程がおかしくなる不具合が確認されています。
  このような場合、 `Ctrl+C` での停止後にArduinoを接続しているUSB端子をいったん抜き差ししてみてください。
- `Sketch version mismatch (…). Please rewrite "bridge/bridge.ino" onto Arduino.` と表示される場合は、ホスト側バイナリとArduino側スケッチのバージョンが一致していません。バイナリを最新版に更新し、スケッチを転送し直す必要があります。
