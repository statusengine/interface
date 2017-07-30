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


class ScheduleddowntimeQueryOptions extends QueryOptions {

    /**
     * @var string
     */
    private $order = 'scheduled_start_time';

    /**
     * @var string
     */
    private $direction = 'ASC';

    /**
     * @var string
     */
    private $hostname__like = '';

    /**
     * @var string
     */
    private $servicedescription__like = '';

    private $is_host_request = true;

    /**
     * ScheduleddowntimeQueryOptions constructor.
     * @param $params
     */
    public function __construct($params) {
        parent::__construct($params);
        $this->params = $params;

        if (isset($this->params['hostname__like'])) {
            $this->hostname__like = $this->params['hostname__like'];
        }

        if (isset($this->params['servicedescription__like'])) {
            $this->servicedescription__like = $this->params['servicedescription__like'];
        }

        $this->is_host_request = true;
        if(isset($this->params['object_type'])){
            if($this->params['object_type'] == 'service'){
                $this->is_host_request = false;
            }
        }

    }

    /**
     * @return string
     */
    public function getHostnameLike() {
        return $this->hostname__like;
    }

    /**
     * @return string
     */
    public function getServicedescriptionLike() {
        return $this->servicedescription__like;
    }

    /**
     * @return string
     */
    public function getOrder() {
        return $this->order;
    }

    /**
     * @return string
     */
    public function getDirection() {
        return $this->direction;
    }

    /**
     * @return bool
     */
    public function isHostRequest(){
        return $this->is_host_request;
    }

}
