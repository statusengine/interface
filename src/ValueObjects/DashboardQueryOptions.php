<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\ValueObjects;


class DashboardQueryOptions extends QueryOptions {

    /**
     * @var array
     */
    private $hostStates = [0,1,2];

    private $serviceStates = [0,1,2,3];

    /**
     * @var bool
     */
    private $excludeAcknowledge = false;

    /**
     * @var bool
     */
    private $excludeInDowntime = false;

    /**
     * DashboardQueryOptions constructor.
     * @param $params
     */
    public function __construct($params) {
        parent::__construct($params);
        if(isset($params['hide_ack_and_downtime']) && $params['hide_ack_and_downtime'] === 'true'){
            $this->excludeAcknowledge = true;
            $this->excludeInDowntime = true;
        }
    }

    /**
     * @return array
     */
    public function getHostStates() {
        return $this->hostStates;
    }

    /**
     * @param array $hostStates
     */
    public function setHostStates($hostStates) {
        $this->hostStates = $hostStates;
    }

    /**
     * @return array
     */
    public function getServiceStates() {
        return $this->serviceStates;
    }

    /**
     * @param array $serviceStates
     */
    public function setServiceStates($serviceStates) {
        $this->serviceStates = $serviceStates;
    }

    /**
     * @return bool
     */
    public function isExcludeAcknowledge() {
        return $this->excludeAcknowledge;
    }

    /**
     * @param bool $excludeAcknowledge
     */
    public function setExcludeAcknowledge($excludeAcknowledge) {
        $this->excludeAcknowledge = $excludeAcknowledge;
    }

    /**
     * @return bool
     */
    public function isExcludeInDowntime() {
        return $this->excludeInDowntime;
    }

    /**
     * @param bool $excludeInDowntime
     */
    public function setExcludeInDowntime($excludeInDowntime) {
        $this->excludeInDowntime = $excludeInDowntime;
    }


    /**
     * @return bool
     */
    public function hasHostStateFilter(){
        if(sizeof($this->hostStates) < 3){
            return true;
        }
        return false;
    }

    /**
     * @return bool
     */
    public function hasServiceStateFilter(){
        if(sizeof($this->serviceStates) < 4){
            return true;
        }
        return false;
    }

}
