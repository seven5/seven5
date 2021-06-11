import basicAttribute from './basicattr';

export default class simpleAttribute<T> extends basicAttribute<T> {
  /**
   * Shorthand for creating a BasicAttribute with a simple value.
   * @param startingValue
   * @param debugName
   */
  constructor(startingValue: T, debugName?: string) {
    super(debugName);
    this.isComputed = false;
    this.dirty = false;
    this.Set(startingValue);
  }
}
