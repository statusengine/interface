<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2018  Daniel Ziegler
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


class ServiceQueryOptions extends QueryOptions {

    private $columnsForOrder = [
        'hostname',
        'service_description',
        'current_state',
        'last_check',
        'output',
        'last_state_change'
    ];

    /**
     * @var string
     */
    private $order = 'hostname, service_description';

    /**
     * @var string
     */
    private $direction = 'asc';

    /**
     * @var array
     */
    private $state = [];

    /**
     * @var array
     */
    private $stateMatch = [
        'ok' => 0,
        'warning' => 1,
        'critical' => 2,
        'unknown' => 3
    ];

    /**
     * @var string
     */
    private $hostname__like = '';

    /**
     * @var string
     */
    private $servicedescription__like = '';

    /**
     * @var string
     */
    private $hostname;

    /**
     * @var string
     */
    private $servicedescription;

    /**
     * @var null|bool
     */
    private $isAcknowledged = null;

    /**
     * @var null|bool
     */
    private $isInDowntime = null;

    /**
     * ServiceQueryOptions constructor.
     * @param $params
     * @throws \Exception
     */
    public function __construct($params) {
        parent::__construct($params);

        if (isset($params['order'])) {
            $params['order'] = explode(',', $params['order']);
            foreach ($params['order'] as $fieldToOrder) {
                if (!in_array($fieldToOrder, $this->columnsForOrder)) {
                    throw new \Exception('Invalid column for order by condition!');
                }
            }
            $this->order = implode(',', $params['order']);
        }

        if (isset($params['direction'])) {
            //asc is class default
            if ($params['direction'] == 'desc') {
                $this->direction = 'desc';
            }
        }

        if (isset($this->params['state']) && is_array($this->params['state'])) {
            $_state = [];
            foreach ($this->params['state'] as $state) {
                if (isset($this->stateMatch[$state])) {
                    $_state[] = $this->stateMatch[$state];
                }
            }
            $this->state = $_state;
        }

        if (isset($this->params['hostname__like'])) {
            $this->hostname__like = $this->params['hostname__like'];
        }

        if (isset($this->params['servicedescription__like'])) {
            $this->servicedescription__like = $this->params['servicedescription__like'];
        }

        if (isset($this->params['hostname'])) {
            $this->hostname = $this->params['hostname'];
        }

        if (isset($this->params['servicedescription'])) {
            $this->servicedescription = $this->params['servicedescription'];
        }

        if (isset($this->params['is_acknowledged'])) {
            $this->isAcknowledged = (bool)$this->params['is_acknowledged'];
        }

        if (isset($this->params['is_in_downtime'])) {
            $this->isInDowntime = (bool)$this->params['is_in_downtime'];
        }

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

    public function getStateFilter() {
        return $this->state;
    }

    public function sizeOfStateFilter() {
        return sizeof($this->state);
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
    public function getHostname() {
        return $this->hostname;
    }

    /**
     * @return string
     */
    public function getServicedescription() {
        return $this->servicedescription;
    }

    /**
     * @return bool|null
     */
    public function getIsAcknowledged() {
        return $this->isAcknowledged;
    }

    /**
     * @return bool|null
     */
    public function getIsInDowntime() {
        return $this->isInDowntime;
    }


}

