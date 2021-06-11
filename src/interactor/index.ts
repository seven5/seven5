import root from './root';
import stickyRoot from './stickyroot';
import interactor from './interactor';
import base from './base';
import attrRect from './attrrect';
import label from './label';
import html from './html';
import column from './column';
import row from './row';
import button from './button';

export {
  /**
   * Default is a base class that most normal interactors will want to use because
   * it provides many sensible defaults, thus the name.
   */
  base as Base,
  /**
   * HTML is an interactor that can render HTML.
   */
  html as HTML,
  /**
   * Interactor is the API to all interactors.  Notationally, I don't
   * like S5Interactor.Interactor which comes from this.  But I do
   * like S5Interactor.Base or S5Interactor.Row, etc.
   */
  interactor as Interactor,
  /**
   * The Root interactor provides all the machinery for handling screen updates.
   * This Root interactor binds itself to the HTML element whose id (in HTML) is
   * matched by the Root interactor property and assumes its
   * the element's size, but doesn't track the element's size.  Use StickyRoot
   * if you want that.
   */
  root as Root,
  /**
   * The StickyRoot interactor is the same as the Root interactor, except for
   * the way it handles DOM changes.  If the element the StickyRoot is bound to
   * in the DOM changes size, StickyRoot makes an identical change to itself.
   */
  stickyRoot as StickyRoot,
  /**
   *  Display a line of text.  Use the flag "fixedSize" and then directly
   *  set the values of W,H if you don't want the label's size to grow to
   *  fit te text.
   */
  label as Label,
  /**
   * Column arranges its children with all the same X value and a gap
   * of margin between children.  Width is width of widest child.
   */
  column as Column,
  /**
   * Row arranges its children with all the same Y value and a gap
   * of margin between children.  Height is height of tallest child.
   */
  row as Row,
  /**
   * Button is a way to capture a click from the user.
   */
  button as Button,
  /**
   * rect made of attributes for each value
   */
  attrRect as AttrRect,
};
