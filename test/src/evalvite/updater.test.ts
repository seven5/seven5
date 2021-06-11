import * as EV from '../../../src/evalvite';
import {expect} from 'chai';

const START_VALUE = -10000;
class testUpdater{
  prev:number;
  current:number;
  attr: EV.Attribute<number> | null;
  conn:EV.Attribute<number> | null;
  disconn: EV.Attribute<number> | null;
  ConnectAttribute(source:EV.Attribute<number>):void {
    this.conn=source;
  }
  DisconnectAttribute(source:EV.Attribute<number>):void {
    this.disconn=source;
  }
  UpdateValue(source:EV.Attribute<number>, newValue: number, oldValue?:number ):void {
    this.attr=source;
    this.current=newValue;
    if (oldValue){
      this.prev=oldValue;
    }
  }
  constructor() {
    this.prev=START_VALUE;
    this.current=START_VALUE;
    this.attr=null;
    this.conn=null;
    this.disconn=null;
  }
}

describe('update', function() {
  it('callback is made', function() {
    const u = new testUpdater();
    const s = EV.Simple<number>(4,"update-callback-constant");
    const c = EV.Computed<number>((x:number)=>x*x, [s],"updater-callback-computed");
    expect(u.conn).to.be.null;
    c.AddUpdateable(u);
    expect(u.conn).to.deep.equal(c);
    expect(u.prev).to.be.equal(START_VALUE);
    expect(u.current).to.be.equal(16);
    expect(c.Get()).to.be.equal(16);
    u.current=17;//safety to make sure no change after next Get()
    expect(c.Get()).to.be.equal(16);//cached
    expect(u.prev).to.be.equal(START_VALUE);
    expect(u.current).to.be.equal(17); //no change
    c.RemoveUpdateable(u);
    expect(u.disconn).to.deep.equal(c);
  });
});