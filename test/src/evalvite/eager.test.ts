import * as EV from '../../../src/evalvite';
import {expect} from 'chai';

const largerGraphHelper = (isEager:boolean):EV.Attribute<number> =>{
}

describe('eager', ()=> {
  it('simplest possible graph', ()=> {
    let count : number = 0;

    const a = EV.Simple<number>(0,"a");
    const b = EV.Computed<number>((n:number):number=>{
      // doing side effects like this is REALLY bad, but ok in a test
      count+=1;
      return n+1;
    },[a],"b")
    b.Eager=true;

    // insure b is up to date
    expect(b.Get()).to.be.equal(1);
    // record evaluations before test
    const before=count;
    // change a
    a.Set(7);
    // we expect that b will be updated, since it is eager
    expect(count).to.be.equal(before+1);
  });
  it('more complex graph', ()=> {
    let count : number = 0;

    const a = EV.Simple<number>(0,"a");
    const b = EV.Computed<number>((n:number):number=>{
      // doing side effects like this is REALLY bad, but ok in a test
      count+=1;
      return n+1;
    },[a],"b")
    const c = EV.Computed<number>((n:number, m:number):number=>{
      // doing side effects like this is REALLY bad, but ok in a test
      count+=1;
      return n+m;
    },[a,b],"c")
    const d = EV.Simple<number>(10,"d");
    const e = EV.Computed<number>((n:number,m:number):number=>{
      // doing side effects like this is REALLY bad, but ok in a test
      count+=1;
      return n+m;
    },[c,d],"e")

    // but a change to a will result in an eager call of e, which evals everything
    e.Eager=true;

    // insure e is up to date
    expect(e.Get()).to.be.equal(11);
    // record evaluations before test
    const before=count;
    // change a
    a.Set(2);
    // we expect that b will be updated, since it is eager
    expect(count).to.be.equal(before+3); //one of b,c, and e
    expect(e.Get()).to.be.equal(15);
    // change d
    d.Set(20);
    expect(count).to.be.equal(before+4); //extra eval of e, but not b,c
    expect(e.Get()).to.be.equal(25);
  });
  it('complex graph, eager in middle', ()=> {
    let count : number = 0;

    const a = EV.Simple<number>(0,"a");
    const b = EV.Computed<number>((n:number):number=>{
      // doing side effects like this is REALLY bad, but ok in a test
      count+=1;
      return n+1;
    },[a],"b")
    const c = EV.Computed<number>((n:number, m:number):number=>{
      // doing side effects like this is REALLY bad, but ok in a test
      count+=1;
      return n+m;
    },[a,b],"c")
    const d = EV.Simple<number>(10,"d");
    const e = EV.Computed<number>((n:number,m:number):number=>{
      // doing side effects like this is REALLY bad, but ok in a test
      count+=1;
      return n+m;
    },[c,d],"e")

    //tricky!
    c.Eager=true;

    // insure e is up to date
    expect(e.Get()).to.be.equal(11);
    // record evaluations before test
    const before=count;
    // change a
    a.Set(2);
    expect(count).to.be.equal(before+2);//e is still lazy and now dirty
    e.Get();
    expect(count).to.be.equal(before+3);
  });
});
