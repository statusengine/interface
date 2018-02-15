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


class ServiceAcknowledgementQueryOptions extends QueryOptions {

    private $columnsForOrder = [
        'hostname',
        'service_description',
        'author_name',
        'comment_data',
        'entry_time',
        'state',
        'acknowledgement_type'
    ];

    /**
     * @var string
     */
    private $order = 'hostname';

    /**
     * @var string
     */
    private $direction = 'asc';

    /**
     * @var array
     */
    private $state = [];

    /**
     * @var int
     */
    private $entry_time__lt = 0;

    /**
     * @var int
     */
    private $entry_time__gt = 0;

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
    private $comment_data__like = '';

    /**
     * @var string
     */
    private $hostname;

    /**
     * @var string
     */
    private $service_description;

    /**
     * @var string
     */
    private $hostname__like = '';

    /**
     * @var string
     */
    private $servicedescription__like = '';

    /**
     * ServiceAcknowledgementQueryOptions constructor.
     * @param $params
     * @throws \Exception
     */
    public function __construct($params) {
        parent::__construct($params);

        if (isset($params['order'])) {
            if (!in_array($params['order'], $this->columnsForOrder)) {
                throw new \Exception('Invalid column for order by condition!');
            }
            $this->order = $params['order'];
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

        if (isset($this->params['hostname'])) {
            $this->hostname = $params['hostname'];
        }

        if (isset($this->params['servicedescription'])) {
            $this->service_description = $params['servicedescription'];
        }

        if (isset($this->params['entry_time__lt'])) {
            $this->entry_time__lt = (int)$this->params['entry_time__lt'];
        }

        if (isset($this->params['entry_time__gt'])) {
            $this->entry_time__gt = (int)$this->params['entry_time__gt'];
        }

        if (isset($this->params['comment_data__like'])) {
            $this->comment_data__like = $this->params['comment_data__like'];
        }

        if (isset($this->params['hostname__like'])) {
            $this->hostname__like = $this->params['hostname__like'];
        }

        if (isset($this->params['servicedescription__like'])) {
            $this->servicedescription__like = $this->params['servicedescription__like'];
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

    /**
     * @return array
     */
    public function getStateFilter() {
        return $this->state;
    }

    /**
     * @return int
     */
    public function sizeOfStateFilter() {
        return sizeof($this->state);
    }

    /**
     * @return string
     */
    public function getCommentDataLike() {
        return $this->comment_data__like;
    }

    /**
     * @return int
     */
    public function getEntryTimeLt() {
        return $this->entry_time__lt;
    }

    /**
     * @return int
     */
    public function getEntryTimeGt() {
        return $this->entry_time__gt;
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
    public function getServiceDescription() {
        return $this->service_description;
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

}

