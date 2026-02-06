-- @author wangcw
-- @copyright (c) 2026, redgreat
-- created : 2026-02-06 15:30:00
-- postgres表结构设计

-- 设置查询路径
alter role user_eadm set search_path to eadm, public;

--设置 本地时区
set time zone 'asia/shanghai';

-- EMQX设备数据表
drop table if exists emqx_device_data cascade;
create table emqx_device_data (
  id serial,
  imei varchar(20) not null,
  lat numeric(10,7),
  lng numeric(10,7),
  gps_ts bigint,
  uptime bigint,
  csq smallint,
  vbat smallint,
  up_vbat smallint,
  ip varchar(50),
  receivetime timestamptz,
  inserttime timestamptz not null default current_timestamp
);

alter table emqx_device_data owner to user_eadm;
alter table emqx_device_data drop constraint if exists pk_emqx_device_data_id cascade;
alter table emqx_device_data add constraint pk_emqx_device_data_id primary key (id);

drop index if exists non_emqx_device_data_imei;
create index non_emqx_device_data_imei on emqx_device_data using btree (imei asc nulls last);
drop index if exists non_emqx_device_data_gps_ts;
create index non_emqx_device_data_gps_ts on emqx_device_data using btree (gps_ts desc nulls last);
drop index if exists non_emqx_device_data_inserttime;
create index non_emqx_device_data_inserttime on emqx_device_data using btree (inserttime desc nulls last);

comment on column emqx_device_data.id is '自增主键';
comment on column emqx_device_data.imei is '设备IMEI号';
comment on column emqx_device_data.lat is 'GPS纬度';
comment on column emqx_device_data.lng is 'GPS经度';
comment on column emqx_device_data.gps_ts is 'GPS时间戳(Unix秒)';
comment on column emqx_device_data.uptime is '设备运行时间(秒)';
comment on column emqx_device_data.csq is '信号质量(0-31)';
comment on column emqx_device_data.vbat is '电池电压(mV)';
comment on column emqx_device_data.up_vbat is '上报时电池电压(mV)';
comment on column emqx_device_data.ip is '设备IP地址';
comment on column emqx_device_data.receivetime is '数据接收时间';
comment on column emqx_device_data.inserttime is '数据写入时间';
comment on table emqx_device_data is 'EMQX设备上报数据表';
