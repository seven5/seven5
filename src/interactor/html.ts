import base from './base';
import interactor from './interactor';
import * as S5Render from '../render';

export default class html extends base implements interactor {
  _html = '';

  img = new Image(); // to get the types right

  loaded = false;

  get Content(): string {
    return this._html;
  }

  set Content(s: string) {
    this._html = s;
    this.loaded = false;
    this.MarkDirty();
  }

  set Height(h: number) {
    super.H = h;
    this.loaded = false;
    this.MarkDirty();
  }

  set Width(w: number) {
    super.W = w;
    this.loaded = false;
    this.MarkDirty();
  }

  Render(ctx: S5Render.Context): void {
    super.RenderNoChildren(ctx);
  }

  RenderSelf(ctx: S5Render.Context): void {
    this.renderHtmlToCanvas(ctx);
    this.MarkClean();
  }

  // hacked version of https://stackoverflow.com/questions/12652769/rendering-html-elements-to-canvas
  renderHtmlToCanvas(ctx: S5Render.Context): void {
    const data =
      `${
        'data:image/svg+xml;charset=utf-8,' +
        '<svg xmlns="http://www.w3.org/2000/svg" width="'
      }${this.W}" height="${this.H}">` +
      `<foreignObject width="100%" height="100%">${this.htmlToXml()}</foreignObject>` +
      `</svg>`;

    if (this.loaded) {
      ctx.DrawImage(this.img, 0, 0);
      return;
    }

    // reset the content
    this.img = new Image();
    this.img.onload = () => {
      this.loaded = true;
      this.MarkDirty();
    };
    this.img.src = data;
  }

  htmlToXml(): string {
    const doc = document.implementation.createHTMLDocument('');
    doc.write(this._html);

    // You must manually set the xmlns if you intend to immediately serialize
    // the HTML document to a string as opposed to appending it to a
    // <foreignObject> in the DOM
    // doc.documentElement.setAttribute('xmlns', doc.documentElement.namespaceURI);

    // Get well-formed markup
    const result = new XMLSerializer().serializeToString(doc.body);
    return result;
  }
}
