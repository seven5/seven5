import interactor from './interactor';
import root from './root';
import * as EV from '../evalvite';
import * as S3Err from '../error';
import { Attribute } from '../evalvite';
import * as S5Conf from '../conf';

let { IdCounter } = S5Conf.Vars;
const { Debug } = S5Conf.Vars;
const { Log } = S5Conf;

export default class stickyRoot extends root implements interactor {
  config = {
    attributes: true,
    // attributeFilter: ['width'],
    // attributeOldValue: true,
    childList: false,
    subtree: false,
  };

  observer: ResizeObserver;

  rawWidth: Attribute<number>;

  rawHeight: Attribute<number>;

  /**
   * Pass in the minimum and maximum values for width and height.  If
   * you pass in zero for the min and max of a dimension, it is unconstrained
   * in size.
   *
   * @param minWidth
   * @param maxWidth
   * @param minHeight
   * @param maxHeight
   * @param debugName
   */
  constructor(
    minWidth: number,
    maxWidth: number,
    minHeight: number,
    maxHeight: number,
    debugName?: string
  ) {
    super();
    if (!debugName) {
      this.dname = `[stickyRoot ${IdCounter}]`;
      IdCounter += 1;
    } else {
      this.dname = debugName;
    }
    this.observer = new ResizeObserver(this.resizeOccurred);
    this.observer.observe(super.Element);

    if (this.Coords.AttrW === null) {
      throw S3Err.NewError(S3Err.Messages.NullAttr, 'width');
    }
    if (this.Coords.AttrH === null) {
      throw S3Err.NewError(S3Err.Messages.NullAttr, 'height');
    }

    this.rawHeight = this.Coords.AttrH;
    this.rawWidth = this.Coords.AttrW;

    if (minWidth !== 0 || maxWidth !== 0) {
      this.Coords.AttrW = EV.MinMax(
        300,
        2400,
        this.rawWidth,
        `${this.dname} minMax width`
      );
    }
    if (minHeight !== 0 || maxHeight !== 0) {
      this.Coords.AttrH = EV.MinMax(
        300,
        1600,
        this.rawHeight,
        `${this.dname} minMax height`
      );
    }
  }

  pushNewDimension(
    isWidth: boolean,
    newValue: number,
    oldValue?: number
  ): void {
    const x = S5Conf.Vars.Debug;
    if (Debug) {
      Log(
        `S5DEBUG:${
          isWidth ? 'width' : 'height'
        } changing from ${oldValue} to ${newValue}
        for element ${this.Element}`
      );
    }
    S5Conf.Vars.Debug = x;
  }

  resizeOccurred = (entries: ResizeObserverEntry[]): void => {
    entries.forEach((ent: ResizeObserverEntry) => {
      if (ent.borderBoxSize && ent.borderBoxSize.length > 0) {
        ent.borderBoxSize.forEach((value: ResizeObserverSize) => {
          const height = value.blockSize;
          const width = value.inlineSize;
          // so the sequence of attributes that are actually used here
          // rawWidth -> minMax -> sideEffect or if you didn't pass min or max
          // rawWidth -> sideEffect
          // and the sideEffect pushes the
          this.rawWidth.Set(width);
          this.rawHeight.Set(height);
          // https://stackoverflow.com/questions/43853119/javascript-wrong-mouse-position-when-drawing-on-canvas
          this.el.style.width = `${this.el.width}px`;
          this.el.style.height = `${this.el.height}px`;
        });
      }
    });
  };

  set Element(el: HTMLCanvasElement) {
    if (this.observer) {
      this.observer.disconnect();
    }
    super.Element = el;
    this.observer.observe(el);
  }
}
