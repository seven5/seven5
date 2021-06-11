import attribute from './attribute';
import basicAttribute from './basicattr';

/**
 * Computed attribute is just a shorthand for creating a basic
 * and assigning it a function.
 */
export default class computedAttribute<T> extends basicAttribute<T> {
  /**
   * This constructs an BasicAttribute whose value is derived from other
   * attributes via fn. See documentation for setting Func.
   * @param fn  given functions must have a precise number of args (no optional or rest arguments)
   * @param inputs
   * @param name
   */
  constructor(
    fn: (...a1: unknown[]) => T,
    inputs: attribute<unknown>[],
    name?: string
  ) {
    super(name);
    this.isComputed = true;
    this.SetInputsAndFunc(inputs, fn);
    // there are no edges yet to notify, so we just mark ourselves
    this.dirty = true;
  }
}
