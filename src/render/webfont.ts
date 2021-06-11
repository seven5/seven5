import { fontStyle, fontWeight } from './font';

export default class webFont {
  baseName = 'josefinslab';

  variableInfo = 'variablefont_wght.';

  _weight: fontWeight;

  _style: fontStyle;

  get Family(): string {
    const style = this.Style === fontStyle.Italic ? '-italic' : '';
    return `${this.baseName}${style}-${this.variableInfo}`;
  }

  get SourceWoff(): string {
    return `${this.Family}-webfont.woff`;
  }

  get SourceWoff2(): string {
    return `${this.Family}-webfont.woff2`;
  }

  get Weight(): fontWeight {
    return this._weight;
  }

  set Weight(w: fontWeight) {
    this._weight = w;
  }

  get Style(): fontStyle {
    return this._style;
  }

  set Style(s: fontStyle) {
    this._style = s;
  }

  variationSetting(): number {
    if (this.Weight === fontWeight.Normal) {
      return 600;
    }
    if (this.Weight === fontWeight.Bold) {
      return 750;
    }
    return 400;
  }

  value(): string {
    return `@font-face {
              font-family: '${this.Family}';
              src:  url('${this.SourceWoff2}') format('woff2'),
                    url('${this.SourceWoff}') format('woff');
              font-weight: ${this.Weight};
              font-style: ${this.Style};
              font-variation-settings: 'wght' ${this.variationSetting()};
            }`;
  }

  constructor() {
    this._weight = fontWeight.Normal;
    this._style = fontStyle.Regular;
  }
}
